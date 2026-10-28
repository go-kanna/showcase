package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-kanna/showcase/internal/errors/aerrors"
	"github.com/go-kanna/showcase/internal/feature/todo/model/fixture"
	"github.com/go-kanna/showcase/internal/feature/todo/postgres"
	"github.com/go-kanna/showcase/internal/feature/todo/usecase"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
	"github.com/go-kanna/showcase/test/infra"
)

func TestDeleteTodo(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	m := fixture.Todo()
	todos := postgres.NewTodo(ifr.Reader)
	ifr.InTx(t, func(tx tx.Tx) {
		_, err := todos.Bind(tx).Upsert(t.Context(), m)
		require.NoError(t, err)
	})
	sut := usecase.NewDeleteTodo(ifr.Infra)

	// act
	_, err := sut.Do(t.Context(), usecase.DeleteTodoInput{TodoID: m.ID.String()})
	_, again := sut.Do(t.Context(), usecase.DeleteTodoInput{TodoID: m.ID.String()})

	// assert
	require.NoError(t, err)
	require.NoError(t, again, "a retry after the todo is gone succeeds")
	got, err := todos.Find(t.Context(), m.ID)
	require.ErrorIs(t, err, aerrors.ErrNotFound)
	assert.Zero(t, got)
}
