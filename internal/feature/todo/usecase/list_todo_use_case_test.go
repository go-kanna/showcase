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
