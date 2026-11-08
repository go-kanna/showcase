package postgres_test

import (
	"errors"
	"testing"
	"time"

	"github.com/go-kanna/orm/option"
	"github.com/go-kanna/value/rule"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-kanna/showcase/internal/errors/aerrors"
	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/model/fixture"
	"github.com/go-kanna/showcase/internal/feature/todo/postgres"
	"github.com/go-kanna/showcase/internal/feature/todo/scopes"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
	"github.com/go-kanna/showcase/test/infra"
)

func TestTodo_GetNotFound(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	sut := postgres.NewTodo(ifr.Reader)

	// act
	_, err := sut.Get(t.Context(), fixture.Todo().ID)

	// assert
	require.ErrorIs(t, err, aerrors.ErrNotFound)
}

func TestTodo_GetForUpdateLocksRow(t *testing.T) {
	t.Parallel()

	ifr := infra.New(t)
	sut := postgres.NewTodo(ifr.Reader)
	m := fixture.Todo()
	ifr.InTx(t, func(tx tx.Tx) {
		require.NoError(t, sut.Bind(tx).Upsert(t.Context(), m))
	})

	ifr.InTx(t, func(tx tx.Tx) {
		got, err := sut.Bind(tx).GetForUpdate(t.Context(), m.ID)
		require.NoError(t, err)
		assert.Equal(t, m.ID, got.ID)

		competing, err := ifr.Writer.Begin(t.Context())
		require.NoError(t, err)
		defer func() { _ = competing.Rollback(t.Context()) }()
		_, err = competing.Exec(t.Context(), "SET LOCAL lock_timeout = '100ms'")
		require.NoError(t, err)

		_, err = competing.Exec(t.Context(), "DELETE FROM todos WHERE id = $1", m.ID)
		pgerr, ok := errors.AsType[*pgconn.PgError](err)
		require.True(t, ok, "err = %v, want a PostgreSQL lock timeout", err)
		assert.Equal(t, "55P03", pgerr.Code)
	})

	_, err := sut.Get(t.Context(), m.ID)
	require.NoError(t, err)
}

func TestTodo_InsertAndGet(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	clock := &infra.Clock{T: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	sut := postgres.NewTodo(ifr.Reader, option.WithClock(clock))
	m := fixture.Todo(func(m *model.Todo) {
		m.CompletedAt = new(model.MustCompletedAt(time.Now()))
	})
	ifr.InTx(t, func(tx tx.Tx) {
		require.NoError(t, sut.Bind(tx).Upsert(t.Context(), m))
	})

	// act
	got, err := sut.Get(t.Context(), m.ID)

	// assert
	require.NoError(t, err)
	assert.Equal(t, m.ID, got.ID)
	assert.Equal(t, m.Title, got.Title)
	assert.Equal(t, m.Priority, got.Priority)
	assert.WithinDuration(t, m.CompletedAt.Time(), got.CompletedAt.Time(), time.Millisecond)
	assert.WithinDuration(t, clock.T, got.CreatedAt, 0, "the repository's clock, not the model's value")
	assert.WithinDuration(t, clock.T, got.UpdatedAt, 0)
}

// A row holding a priority the model has no constant for, as one written by a
// newer release and read after a rollback, is an error rather than a to-do
// holding it.
func TestTodo_GetUnknownPriority(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	sut := postgres.NewTodo(ifr.Reader)
	m := fixture.Todo()
	ifr.InTx(t, func(tx tx.Tx) {
		require.NoError(t, sut.Bind(tx).Upsert(t.Context(), m))
		_, err := tx.DBTX().Exec(t.Context(), "UPDATE todos SET priority = 'urgent' WHERE id = $1", m.ID)
		require.NoError(t, err)
	})

	// act
	_, err := sut.Get(t.Context(), m.ID)

	// assert
	rerr, ok := errors.AsType[*rule.Error](err)
	require.True(t, ok, "err = %v, want a *rule.Error", err)
	assert.Equal(t, "inclusion", rerr.Code())
}

func TestTodo_UpdateAndGet(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	clock := &infra.Clock{T: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	sut := postgres.NewTodo(ifr.Reader, option.WithClock(clock))
	m := fixture.Todo()
	ifr.InTx(t, func(tx tx.Tx) {
		require.NoError(t, sut.Bind(tx).Upsert(t.Context(), m))
	})
	created := clock.T
	clock.T = clock.T.Add(time.Hour)
	m.CompletedAt = new(model.MustCompletedAt(clock.T))

	// act
	ifr.InTx(t, func(tx tx.Tx) {
		require.NoError(t, sut.Bind(tx).Upsert(t.Context(), m))
	})

	// assert
	got, err := sut.Get(t.Context(), m.ID)
	require.NoError(t, err)
	assert.WithinDuration(t, clock.T, got.CompletedAt.Time(), 0)
	assert.WithinDuration(t, created, got.CreatedAt, 0, "an upsert of a row there is leaves its created_at")
	assert.WithinDuration(t, clock.T, got.UpdatedAt, 0)
}

// Update writes the columns of the change set and the time of the update, and
// leaves the others as they were.
func TestTodo_Update(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	clock := &infra.Clock{T: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	sut := postgres.NewTodo(ifr.Reader, option.WithClock(clock))
	m := fixture.Todo()
	ifr.InTx(t, func(tx tx.Tx) {
		require.NoError(t, sut.Bind(tx).Upsert(t.Context(), m))
	})
	created := clock.T
	clock.T = clock.T.Add(time.Hour)
	c := scopes.TodoChanges().SetCompletedAt(new(model.MustCompletedAt(clock.T)))

	// act
	ifr.InTx(t, func(tx tx.Tx) {
		require.NoError(t, sut.Bind(tx).Update(t.Context(), m.ID, c))
	})

	// assert
	got, err := sut.Get(t.Context(), m.ID)
	require.NoError(t, err)
	require.NotNil(t, got.CompletedAt)
	assert.WithinDuration(t, clock.T, got.CompletedAt.Time(), 0)
	assert.WithinDuration(t, clock.T, got.UpdatedAt, 0)
	assert.WithinDuration(t, created, got.CreatedAt, 0)
	assert.Equal(t, m.Title, got.Title)
}

func TestTodo_UpdateNotFound(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	sut := postgres.NewTodo(ifr.Reader)

	// act
	var err error
	ifr.InTx(t, func(tx tx.Tx) {
		err = sut.Bind(tx).Update(t.Context(), fixture.Todo().ID, scopes.TodoChanges().SetTitle("renamed"))
	})

	// assert
	require.ErrorIs(t, err, aerrors.ErrNotFound)
}

func TestTodo_Delete(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	sut := postgres.NewTodo(ifr.Reader)
	m := fixture.Todo()
	ifr.InTx(t, func(tx tx.Tx) {
		require.NoError(t, sut.Bind(tx).Upsert(t.Context(), m))
	})

	// act
	ifr.InTx(t, func(tx tx.Tx) {
		require.NoError(t, sut.Bind(tx).Delete(t.Context(), m.ID))
	})

	// assert
	got, err := sut.Get(t.Context(), m.ID)
	require.ErrorIs(t, err, aerrors.ErrNotFound)
	assert.Zero(t, got)
}

func TestTodo_DeleteNotFound(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	sut := postgres.NewTodo(ifr.Reader)

	// act
	var err error
	ifr.InTx(t, func(tx tx.Tx) {
		err = sut.Bind(tx).Delete(t.Context(), fixture.Todo().ID)
	})

	// assert
	require.ErrorIs(t, err, aerrors.ErrNotFound)
}
