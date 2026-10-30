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

func TestPutTodo_Insert(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	todos := postgres.NewTodo(ifr.Reader)
	m := fixture.Todo()
	completedAt := time.Now()

	// act
	_, err := usecase.NewPutTodo(ifr.Transactor, todos).Do(t.Context(), usecase.PutTodoInput{
		ID:          m.ID.String(),
		Title:       m.Title,
		Priority:    model.PriorityHigh,
		CompletedAt: &completedAt,
	})

	// assert
	require.NoError(t, err)
	got, err := todos.Get(t.Context(), m.ID)
	require.NoError(t, err)
	assert.Equal(t, m.ID, got.ID)
	assert.Equal(t, m.Title, got.Title)
	assert.Equal(t, model.PriorityHigh, got.Priority)
	require.NotNil(t, got.CompletedAt)
	assert.WithinDuration(t, completedAt, got.CompletedAt.Time(), time.Millisecond)
	assert.WithinDuration(t, time.Now(), got.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), got.UpdatedAt, time.Second)
}

func TestPutTodo_Update(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	todos := postgres.NewTodo(ifr.Reader)
	m := fixture.Todo(func(m *model.Todo) {
		m.CompletedAt = new(model.MustCompletedAt(time.Now()))
	})
	ifr.InTx(t, func(tx tx.Tx) {
		err := todos.Bind(tx).Upsert(t.Context(), m)
		require.NoError(t, err)
	})
	title := m.Title + " (updated)"

	// act
	_, err := usecase.NewPutTodo(ifr.Transactor, todos).Do(t.Context(), usecase.PutTodoInput{
		ID:       m.ID.String(),
		Title:    title,
		Priority: m.Priority,
	})

	// assert
	require.NoError(t, err)
	got, err := todos.Get(t.Context(), m.ID)
	require.NoError(t, err)
	assert.Equal(t, m.ID, got.ID)
	assert.Equal(t, title, got.Title)
	assert.Nil(t, got.CompletedAt)
	assert.WithinDuration(t, m.CreatedAt, got.CreatedAt, time.Millisecond)
	assert.WithinDuration(t, time.Now(), got.UpdatedAt, time.Second)
}

// The rest need no database: the input is checked before any repository is
// called, and what reaches one is told by a fake.

func TestPutTodo_Upserts(t *testing.T) {
	t.Parallel()

	// arrange
	todos := &fakeTodos{}
	m := fixture.Todo()

	// act
	_, err := usecase.NewPutTodo(fakeTransactor{}, todos).Do(t.Context(), usecase.PutTodoInput{
		ID:       m.ID.String(),
		Title:    m.Title,
		Priority: model.PriorityHigh,
	})

	// assert
	require.NoError(t, err)
	require.Len(t, todos.upserted, 1)
	got := todos.upserted[0]
	assert.Equal(t, m.ID, got.ID)
	assert.Equal(t, m.Title, got.Title)
	assert.Equal(t, model.PriorityHigh, got.Priority)
	assert.Nil(t, got.CompletedAt)
	assert.Equal(t, got.CreatedAt, got.UpdatedAt)
}

func TestPutTodo_InvalidCompletedAt(t *testing.T) {
	t.Parallel()

	// arrange
	todos := &fakeTodos{}
	m := fixture.Todo()
	completedAt := time.Date(1999, time.December, 31, 0, 0, 0, 0, time.UTC)

	// act
	_, err := usecase.NewPutTodo(fakeTransactor{}, todos).Do(t.Context(), usecase.PutTodoInput{
		ID:          m.ID.String(),
		Title:       m.Title,
		Priority:    m.Priority,
		CompletedAt: &completedAt,
	})

	// assert
	var verrs validate.Errors
	require.ErrorAs(t, err, &verrs)
	require.Len(t, verrs, 1)
	assert.Equal(t, "completed_at", verrs[0].Field)
	assert.Equal(t, "greater_than_or_equal_to", verrs[0].Code())
	assert.Empty(t, todos.upserted)
}

// A request without a priority comes in as PRIORITY_UNSPECIFIED, which the
// mapper reads as the zero value, the one that stands for none.
func TestPutTodo_NoPriority(t *testing.T) {
	t.Parallel()

	// arrange
	todos := &fakeTodos{}
	m := fixture.Todo()

	// act
	_, err := usecase.NewPutTodo(fakeTransactor{}, todos).Do(t.Context(), usecase.PutTodoInput{
		ID:    m.ID.String(),
		Title: m.Title,
	})

	// assert
	var verrs validate.Errors
	require.ErrorAs(t, err, &verrs)
	require.Len(t, verrs, 1)
	assert.Equal(t, "priority", verrs[0].Field)
	assert.Equal(t, "blank", verrs[0].Code())
	assert.Empty(t, todos.upserted)
}
