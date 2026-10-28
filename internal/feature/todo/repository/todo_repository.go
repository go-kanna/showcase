package repository

import (
	"context"
	"time"
	"uuid"

	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
)

type Todo interface {
	Find(ctx context.Context, id uuid.UUID) (model.Todo, error)
	List(ctx context.Context) ([]model.Todo, error)
	Upsert(ctx context.Context, m model.Todo) (bool, error)
	Complete(ctx context.Context, id uuid.UUID, at time.Time) (bool, error)
	Uncomplete(ctx context.Context, id uuid.UUID) (bool, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Bind(tx tx.Tx) Todo
}
