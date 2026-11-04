package repository

import (
	"context"
	"time"
	"uuid"

	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/scopes"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
)

//go:generate go tool kanna-orm -destination ../postgres

//kanna:repository
type Todo interface {
	Get(ctx context.Context, id uuid.UUID) (model.Todo, error)
	//kanna:lock update
	GetForUpdate(ctx context.Context, id uuid.UUID) (model.Todo, error)
	// ListWithCommentCounts runs the query of its name in todo_queries.sql, and
	// narrows, orders, and pages its rows by the scopes, as those of todos.
	ListWithCommentCounts(ctx context.Context, s ...scopes.Todo) ([]TodoWithCommentCount, error)
	//kanna:conflict id
	Upsert(ctx context.Context, m model.Todo) error
	Complete(ctx context.Context, id uuid.UUID, completedAt time.Time) (int64, error)
	Uncomplete(ctx context.Context, id uuid.UUID, updatedAt time.Time) (int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Bind(tx tx.Tx) Todo
}

// TodoWithCommentCount is a to-do with how many comments it has, which the
// query adds to its row. The count stays out of model.Todo, which every other
// method reads without it.
type TodoWithCommentCount struct {
	model.Todo

	CommentCount int64
}
