package usecase_test

import (
	"testing"
	"time"
	"uuid"

	"github.com/go-kanna/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-kanna/showcase/internal/errors/aerrors"
	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/model/fixture"
	"github.com/go-kanna/showcase/internal/feature/todo/postgres"
	"github.com/go-kanna/showcase/internal/feature/todo/projection"
	"github.com/go-kanna/showcase/internal/feature/todo/usecase"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
	"github.com/go-kanna/showcase/internal/lib/xslices"
	"github.com/go-kanna/showcase/internal/messages"
	"github.com/go-kanna/showcase/test/infra"
)

func TestListTodo(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	m1 := fixture.Todo(func(m *model.Todo) {
		m.CompletedAt = new(model.MustCompletedAt(time.Now()))
	})
	m2 := fixture.Todo(func(m *model.Todo) {
		m.CompletedAt = new(model.MustCompletedAt(time.Now()))
	})
	for _, m := range []model.Todo{m1, m2} {
		ifr.InTx(t, func(tx tx.Tx) {
			err := postgres.NewTodo(ifr.Reader).Bind(tx).Upsert(t.Context(), m)
			require.NoError(t, err)
		})
	}

	// act
	sut := usecase.NewListTodo(ifr.ReadTransactor, postgres.NewTodo(ifr.Reader))
	got, err := sut.Do(t.Context(), usecase.ListTodoInput{})

	// assert
	require.NoError(t, err)
	assert.Len(t, got.Todos, 2)
	assert.Equal(t, m1.ID, got.Todos[0].ID)
	assert.Equal(t, m2.ID, got.Todos[1].ID)
}

// A filter narrows the list to the to-dos it names, one variant of the sum
// type at a time.
func TestListTodo_Filter(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	done := fixture.Todo(func(m *model.Todo) {
		m.Priority = model.PriorityHigh
		m.CompletedAt = new(model.MustCompletedAt(time.Now()))
	})
	open := fixture.Todo(func(m *model.Todo) { m.Priority = model.PriorityLow })
	for _, m := range []model.Todo{done, open} {
		ifr.InTx(t, func(tx tx.Tx) {
			require.NoError(t, postgres.NewTodo(ifr.Reader).Bind(tx).Upsert(t.Context(), m))
		})
	}
	sut := usecase.NewListTodo(ifr.ReadTransactor, postgres.NewTodo(ifr.Reader))

	tests := []struct {
		name   string
		filter usecase.ListFilter
		want   []model.Todo
	}{
		{name: "completed", filter: usecase.Completion{Completed: true}, want: []model.Todo{done}},
		{name: "not completed", filter: usecase.Completion{}, want: []model.Todo{open}},
		{name: "of a priority", filter: usecase.Priority{Priority: model.PriorityLow}, want: []model.Todo{open}},
		{name: "none", want: []model.Todo{done, open}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// act
			got, err := sut.Do(t.Context(), usecase.ListTodoInput{Filter: tt.filter})

			// assert
			require.NoError(t, err)
			require.Len(t, got.Todos, len(tt.want))
			for i, m := range tt.want {
				assert.Equal(t, m.ID, got.Todos[i].ID)
			}
		})
	}
}

// A priority filter left unspecified is reported at its path, before any
// repository is called.
func TestListTodo_FilterWithoutPriority(t *testing.T) {
	t.Parallel()

	// act
	_, err := usecase.NewListTodo(nil, nil).Do(t.Context(), usecase.ListTodoInput{Filter: usecase.Priority{}})

	// assert
	var verrs validator.Errors
	require.ErrorAs(t, err, &verrs)
	require.Len(t, verrs, 1)
	assert.Equal(t, "filter.priority", verrs[0].Field)
	assert.Equal(t, "blank", verrs[0].Code())
}

func TestListTodo_FilterWithUnknownPriority(t *testing.T) {
	t.Parallel()

	_, err := usecase.NewListTodo(nil, nil).Do(t.Context(), usecase.ListTodoInput{
		Filter: usecase.Priority{Priority: model.Priority("urgent")},
	})

	var verrs validator.Errors
	require.ErrorAs(t, err, &verrs)
	require.Len(t, verrs, 1)
	assert.Equal(t, "filter.priority", verrs[0].Field)
	assert.Equal(t, "inclusion", verrs[0].Code())
}

// A pointer to a variant implements ListFilter too, but Do narrows by the
// variants alone, so it is refused before any repository is called, rather than
// listing every to-do.
func TestListTodo_PointerFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		filter usecase.ListFilter
	}{
		{name: "to a variant", filter: &usecase.Priority{Priority: model.PriorityLow}},
		{name: "nil", filter: (*usecase.Completion)(nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// act
			_, err := usecase.NewListTodo(nil, nil).Do(t.Context(), usecase.ListTodoInput{Filter: tt.filter})

			// assert
			require.ErrorContains(t, err, "unsupported filter")
		})
	}
}

// A page goes on from the token of the page before, and the last page gives
// none.
func TestListTodo_Pages(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	ms := []model.Todo{fixture.Todo(), fixture.Todo(), fixture.Todo()}
	for _, m := range ms {
		ifr.InTx(t, func(tx tx.Tx) {
			require.NoError(t, postgres.NewTodo(ifr.Reader).Bind(tx).Upsert(t.Context(), m))
		})
	}
	sut := usecase.NewListTodo(ifr.ReadTransactor, postgres.NewTodo(ifr.Reader))

	// act
	first, err := sut.Do(t.Context(), usecase.ListTodoInput{PageSize: 2})
	require.NoError(t, err)
	second, err := sut.Do(t.Context(), usecase.ListTodoInput{PageSize: 2, PageToken: first.NextPageToken})
	require.NoError(t, err)

	// assert
	require.Len(t, first.Todos, 2)
	assert.Equal(t, ms[0].ID, first.Todos[0].ID)
	assert.Equal(t, ms[1].ID, first.Todos[1].ID)
	assert.NotEmpty(t, first.NextPageToken)
	require.Len(t, second.Todos, 1)
	assert.Equal(t, ms[2].ID, second.Todos[0].ID)
	assert.Empty(t, second.NextPageToken)
}

// A token the service did not give is refused before any repository is called.
func TestListTodo_InvalidPageToken(t *testing.T) {
	t.Parallel()

	// act
	_, err := usecase.NewListTodo(nil, nil).Do(t.Context(), usecase.ListTodoInput{PageToken: "not a token"})

	// assert
	require.ErrorIs(t, err, aerrors.ErrInvalidArgument)
	var aerr aerrors.Error
	require.ErrorAs(t, err, &aerr)
	m, ok := aerr.Message()
	require.True(t, ok)
	assert.Equal(t, messages.TodoInvalidPageToken(), m)
}

func TestListTodo_PageSizeOutOfRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		size int32
		code string
	}{
		{size: -1, code: "greater_than_or_equal_to"},
		{size: 101, code: "less_than_or_equal_to"},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			t.Parallel()

			// act
			_, err := usecase.NewListTodo(nil, nil).Do(t.Context(), usecase.ListTodoInput{PageSize: tt.size})

			// assert
			var verrs validator.Errors
			require.ErrorAs(t, err, &verrs)
			require.Len(t, verrs, 1)
			assert.Equal(t, "page_size", verrs[0].Field)
			assert.Equal(t, tt.code, verrs[0].Code())
		})
	}
}

// Each to-do comes with its own latest comments, newest first: the limit keeps
// a page of the comments of each to-do, not of the list.
func TestListTodo_LatestComments(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	busy, quiet, silent := fixture.Todo(), fixture.Todo(), fixture.Todo()
	var comments []model.Comment // as inserted, with the IDs the database gave
	ifr.InTx(t, func(tx tx.Tx) {
		for _, m := range []model.Todo{busy, quiet, silent} {
			require.NoError(t, postgres.NewTodo(ifr.Reader).Bind(tx).Upsert(t.Context(), m))
		}
		for _, todoID := range []uuid.UUID{busy.ID, busy.ID, busy.ID, busy.ID, quiet.ID} {
			c, err := postgres.NewComment(ifr.Reader).Bind(tx).
				Insert(t.Context(), fixture.Comment(func(c *model.Comment) { c.TodoID = todoID }))
			require.NoError(t, err)
			comments = append(comments, c)
		}
	})
	sut := usecase.NewListTodo(ifr.ReadTransactor, postgres.NewTodo(ifr.Reader))

	// act
	got, err := sut.Do(t.Context(), usecase.ListTodoInput{})

	// assert
	require.NoError(t, err)
	require.Len(t, got.Todos, 3)
	ids := func(cs []model.Comment) []uuid.UUID {
		return xslices.Map(cs, func(c model.Comment) uuid.UUID { return c.ID })
	}
	assert.Equal(t, []uuid.UUID{comments[3].ID, comments[2].ID, comments[1].ID}, ids(got.Todos[0].Comments))
	assert.Equal(t, []uuid.UUID{comments[4].ID}, ids(got.Todos[1].Comments))
	assert.NotNil(t, got.Todos[2].Comments, "loaded, with no comments")
	assert.Empty(t, got.Todos[2].Comments)
	// The count, which the query written by hand adds, counts them all.
	assert.Equal(t, []int64{4, 1, 0}, xslices.Map(got.Todos, func(t projection.TodoWithCommentCount) int64 {
		return t.CommentCount
	}))
}
