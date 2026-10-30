package usecase_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
