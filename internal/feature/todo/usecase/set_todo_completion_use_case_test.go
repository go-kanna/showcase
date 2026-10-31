package usecase_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-kanna/showcase/internal/errors/aerrors"
	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/model/fixture"
	"github.com/go-kanna/showcase/internal/feature/todo/postgres"
	"github.com/go-kanna/showcase/internal/feature/todo/usecase"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
	"github.com/go-kanna/showcase/test/infra"
)

func TestSetTodoCompletion_Complete(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	m := fixture.Todo()
	todos := postgres.NewTodo(ifr.Reader)
	ifr.InTx(t, func(tx tx.Tx) {
		err := todos.Bind(tx).Upsert(t.Context(), m)
		require.NoError(t, err)
	})

	// act
	sut := usecase.NewSetTodoCompletion(ifr.Transactor, todos)
	_, err := sut.Do(t.Context(), usecase.SetTodoCompletionInput{
		TodoID:    m.ID.String(),
		Completed: true,
	})

	// assert
	require.NoError(t, err)
	got, err := todos.Get(t.Context(), m.ID)
	require.NoError(t, err)
	require.NotNil(t, got.CompletedAt)
	assert.WithinDuration(t, time.Now(), got.CompletedAt.Time(), time.Second)
}

func TestSetTodoCompletion_Uncomplete(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	m := fixture.Todo(func(m *model.Todo) {
		m.CompletedAt = new(model.MustCompletedAt(time.Now()))
	})
	todos := postgres.NewTodo(ifr.Reader)
	ifr.InTx(t, func(tx tx.Tx) {
		err := todos.Bind(tx).Upsert(t.Context(), m)
		require.NoError(t, err)
	})

	// act
	sut := usecase.NewSetTodoCompletion(ifr.Transactor, todos)
	_, err := sut.Do(t.Context(), usecase.SetTodoCompletionInput{
		TodoID:    m.ID.String(),
		Completed: false,
	})

	// assert
	require.NoError(t, err)
	got, err := todos.Get(t.Context(), m.ID)
	require.NoError(t, err)
	assert.Nil(t, got.CompletedAt)
}

func TestSetTodoCompletion_NotFound(t *testing.T) {
	t.Parallel()

	ifr := infra.New(t)
	todos := postgres.NewTodo(ifr.Reader)
	_, err := usecase.NewSetTodoCompletion(ifr.Transactor, todos).Do(t.Context(), usecase.SetTodoCompletionInput{
		TodoID:    fixture.Todo().ID.String(),
		Completed: true,
	})

	require.ErrorIs(t, err, aerrors.ErrNotFound)
}

func TestSetTodoCompletion_AlreadyInState(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name      string
		completed bool
	}{
		{name: "completed", completed: true},
		{name: "not completed"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ifr := infra.New(t)
			m := fixture.Todo(func(m *model.Todo) {
				if tt.completed {
					m.CompletedAt = new(model.MustCompletedAt(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))
				}
			})
			todos := postgres.NewTodo(ifr.Reader)
			ifr.InTx(t, func(tx tx.Tx) {
				require.NoError(t, todos.Bind(tx).Upsert(t.Context(), m))
			})

			_, err := usecase.NewSetTodoCompletion(ifr.Transactor, todos).Do(t.Context(), usecase.SetTodoCompletionInput{
				TodoID:    m.ID.String(),
				Completed: tt.completed,
			})

			require.NoError(t, err)
			got, err := todos.Get(t.Context(), m.ID)
			require.NoError(t, err)
			if tt.completed {
				require.NotNil(t, got.CompletedAt)
				assert.WithinDuration(t, m.CompletedAt.Time(), got.CompletedAt.Time(), 0)
			} else {
				assert.Nil(t, got.CompletedAt)
			}
		})
	}
}
