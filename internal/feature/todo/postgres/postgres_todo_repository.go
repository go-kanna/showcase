package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"

	"github.com/go-kanna/showcase/internal/errors/aerrors"
	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/queries"
	"github.com/go-kanna/showcase/internal/feature/todo/repository"
	"github.com/go-kanna/showcase/internal/infra/storage/db"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
	"github.com/go-kanna/showcase/internal/lib/ptr"
)

type Todo struct {
	q *queries.Queries
}

var _ repository.Todo = &Todo{}

func NewTodo(reader db.Reader) repository.Todo {
	return &Todo{q: queries.New(reader)}
}

func (r *Todo) Bind(tx tx.Tx) repository.Todo {
	return &Todo{q: queries.New(tx.DBTX())}
}

func (r *Todo) Find(ctx context.Context, id uuid.UUID) (model.Todo, error) {
	row, err := r.q.GetTodo(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Todo{}, aerrors.NotFound("todo")
	}
	if err != nil {
		return model.Todo{}, fmt.Errorf("get todo: %w", err)
	}
	m, err := queries.RowToTodo(row)
	if err != nil {
		return model.Todo{}, fmt.Errorf("get todo: %w", err)
	}
	return m, nil
}

func (r *Todo) List(ctx context.Context) ([]model.Todo, error) {
	rows, err := r.q.ListTodo(ctx)
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}
	todos := make([]model.Todo, len(rows))
	for i, row := range rows {
		todos[i], err = queries.RowToTodo(row)
		if err != nil {
			return nil, fmt.Errorf("list todos: %w", err)
		}
	}
	return todos, nil
}

func (r *Todo) Upsert(ctx context.Context, m model.Todo) (bool, error) {
	n, err := r.q.UpsertTodo(ctx, queries.UpsertTodoParams{
		ID:    m.ID,
		Title: m.Title,
		CompletedAt: ptr.Map(m.CompletedAt, model.CompletedAt.Time),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	})
	if err != nil {
		return false, fmt.Errorf("upsert todo: %w", err)
	}
	return n > 0, nil
}

func (r *Todo) Complete(
	ctx context.Context,
	id uuid.UUID,
	at time.Time,
) (bool, error) {
	n, err := r.q.CompleteTodo(ctx, queries.CompleteTodoParams{
		CompletedAt: at,
		ID:          id,
	})
	if err != nil {
		return false, fmt.Errorf("complete todo: %w", err)
	}
	if n == 0 {
		return false, nil
	}
	return true, nil
}

func (r *Todo) Uncomplete(
	ctx context.Context,
	id uuid.UUID,
) (bool, error) {
	n, err := r.q.UncompleteTodo(ctx, id)
	if err != nil {
		return false, fmt.Errorf("uncomplete todo: %w", err)
	}
	return n > 0, nil
}

func (r *Todo) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	if err := r.q.DeleteTodo(ctx, id); err != nil {
		return fmt.Errorf("delete todo: %w", err)
	}
	return nil
}
