package usecase_test

import (
	"testing"
	"time"

	"github.com/go-kanna/validate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/model/fixture"
	"github.com/go-kanna/showcase/internal/feature/todo/postgres"
	"github.com/go-kanna/showcase/internal/feature/todo/usecase"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
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
		{name: "completed", filter: usecase.CompletionFilter{Completed: true}, want: []model.Todo{done}},
		{name: "not completed", filter: usecase.CompletionFilter{}, want: []model.Todo{open}},
		{name: "of a priority", filter: usecase.PriorityFilter{Priority: model.PriorityLow}, want: []model.Todo{open}},
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
	_, err := usecase.NewListTodo(nil, nil).Do(t.Context(), usecase.ListTodoInput{Filter: usecase.PriorityFilter{}})

	// assert
	var verrs validate.Errors
	require.ErrorAs(t, err, &verrs)
	require.Len(t, verrs, 1)
	assert.Equal(t, "filter.priority", verrs[0].Field)
	assert.Equal(t, "blank", verrs[0].Code())
}

func TestListTodo_FilterWithUnknownPriority(t *testing.T) {
	t.Parallel()

	_, err := usecase.NewListTodo(nil, nil).Do(t.Context(), usecase.ListTodoInput{
		Filter: usecase.PriorityFilter{Priority: model.Priority("urgent")},
	})

	var verrs validate.Errors
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
		{name: "to a variant", filter: &usecase.PriorityFilter{Priority: model.PriorityLow}},
		{name: "nil", filter: (*usecase.CompletionFilter)(nil)},
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
