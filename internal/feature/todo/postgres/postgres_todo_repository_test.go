package postgres_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-kanna/showcase/internal/errors/aerrors"
	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/model/fixture"
	"github.com/go-kanna/showcase/internal/feature/todo/postgres"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
	"github.com/go-kanna/showcase/test/infra"
)

func TestTodo_FindNotFound(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	sut := postgres.NewTodo(ifr.Reader)

	// act
	_, err := sut.Find(t.Context(), fixture.Todo().ID)

	// assert
	require.ErrorIs(t, err, aerrors.ErrNotFound)
}

func TestTodo_InsertAndFind(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	sut := postgres.NewTodo(ifr.Reader)
	m := fixture.Todo(func(m *model.Todo) {
		m.CompletedAt = new(model.MustCompletedAt(time.Now()))
	})
	ifr.InTx(t, func(tx tx.Tx) {
		_, err := sut.Bind(tx).Upsert(t.Context(), m)
		require.NoError(t, err)
	})

	// act
	got, err := sut.Find(t.Context(), m.ID)

	// assert
	require.NoError(t, err)
	assert.Equal(t, m.ID, got.ID)
	assert.Equal(t, m.Title, got.Title)
	assert.WithinDuration(t, m.CompletedAt.Time(), got.CompletedAt.Time(), time.Millisecond)
	assert.WithinDuration(t, m.CreatedAt, got.CreatedAt, time.Millisecond)
	assert.WithinDuration(t, m.UpdatedAt, got.UpdatedAt, time.Millisecond)
}

func TestTodo_UpdateAndFind(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	sut := postgres.NewTodo(ifr.Reader)
	m := fixture.Todo()
	ifr.InTx(t, func(tx tx.Tx) {
		_, err := sut.Bind(tx).Upsert(t.Context(), m)
		require.NoError(t, err)
	})
	now := time.Now()
	m.CompletedAt = new(model.MustCompletedAt(now))
	m.UpdatedAt = now

	// act
	ifr.InTx(t, func(tx tx.Tx) {
		updated, err := sut.Bind(tx).Upsert(t.Context(), m)
		require.NoError(t, err)
		assert.True(t, updated)
	})

	// assert
	got, err := sut.Find(t.Context(), m.ID)
	require.NoError(t, err)
	assert.WithinDuration(t, now, got.CompletedAt.Time(), time.Millisecond)
	assert.WithinDuration(t, now, got.UpdatedAt, time.Millisecond)
}

func TestTodo_Complete(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	sut := postgres.NewTodo(ifr.Reader)
	m := fixture.Todo()
	ifr.InTx(t, func(tx tx.Tx) {
		_, err := sut.Bind(tx).Upsert(t.Context(), m)
		require.NoError(t, err)
	})
	now := time.Now()

	// act
	ifr.InTx(t, func(tx tx.Tx) {
		updated, err := sut.Bind(tx).Complete(t.Context(), m.ID, now)
		require.NoError(t, err)
		assert.True(t, updated)
	})

	// assert
	got, err := sut.Find(t.Context(), m.ID)
	require.NoError(t, err)
	assert.WithinDuration(t, now, got.CompletedAt.Time(), time.Millisecond)
}

func TestTodo_Uncomplete(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	sut := postgres.NewTodo(ifr.Reader)
	m := fixture.Todo(func(m *model.Todo) {
		m.CompletedAt = new(model.MustCompletedAt(time.Now()))
	})
	ifr.InTx(t, func(tx tx.Tx) {
		_, err := sut.Bind(tx).Upsert(t.Context(), m)
		require.NoError(t, err)
	})

	// act
	ifr.InTx(t, func(tx tx.Tx) {
		updated, err := sut.Bind(tx).Uncomplete(t.Context(), m.ID)
		require.NoError(t, err)
		assert.True(t, updated)
	})

	// assert
	got, err := sut.Find(t.Context(), m.ID)
	require.NoError(t, err)
	assert.Nil(t, got.CompletedAt)
}

func TestTodo_Delete(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	sut := postgres.NewTodo(ifr.Reader)
	m := fixture.Todo()
	ifr.InTx(t, func(tx tx.Tx) {
		_, err := sut.Bind(tx).Upsert(t.Context(), m)
		require.NoError(t, err)
	})

	// act
	ifr.InTx(t, func(tx tx.Tx) {
		require.NoError(t, sut.Bind(tx).Delete(t.Context(), m.ID))
	})

	// assert
	got, err := sut.Find(t.Context(), m.ID)
	require.ErrorIs(t, err, aerrors.ErrNotFound)
	assert.Zero(t, got)
}
