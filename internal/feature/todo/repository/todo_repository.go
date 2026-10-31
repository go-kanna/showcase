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

//kanna:repository table=todos
type Todo interface {
	Get(ctx context.Context, id uuid.UUID) (model.Todo, error)
	GetForUpdate(ctx context.Context, id uuid.UUID) (model.Todo, error)
	List(ctx context.Context, s ...scopes.Todo) ([]model.Todo, error)
	//kanna:conflict id keep=created_at
	Upsert(ctx context.Context, m model.Todo) error
	Complete(ctx context.Context, id uuid.UUID, completedAt time.Time) (int64, error)
	Uncomplete(ctx context.Context, id uuid.UUID) (int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Bind(tx tx.Tx) Todo
}
