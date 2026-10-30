package usecase

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/repository"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
	"github.com/go-kanna/showcase/internal/lib/either"
	"github.com/go-kanna/showcase/internal/lib/ptr"
)

//kanna:validate
type PutTodoInput struct {
	ID          string         `json:"id"           validate:"required,uuid"`
	Title       string         `json:"title"        validate:"required,min=1,max=200"`
	Priority    model.Priority `json:"priority"     validate:"required"`
	CompletedAt *time.Time     `json:"completed_at" validate:"rule=model.CompletedAt"`
}

type PutTodoOutput struct {
}

//kanna:container
type PutTodo struct {
	transactor tx.Transactor   `di:"arg"`
	todos      repository.Todo `di:"arg"`
}

func (uc PutTodo) Do(ctx context.Context, in PutTodoInput) (PutTodoOutput, error) {
	if err := in.Validate(); err != nil {
		return PutTodoOutput{}, err
	}

	id := either.Must(uuid.Parse(in.ID))
	if err := uc.transactor.WithTx(ctx, func(tx tx.Tx) error {
		// Upsert keeps the created_at of a to-do there is, as its //kanna:conflict says, so nothing needs reading first.
		now := time.Now()
		m := model.Todo{
			ID:          id,
			Title:       in.Title,
			Priority:    in.Priority,
			CompletedAt: ptr.Map(in.CompletedAt, model.MustCompletedAt),
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := uc.todos.Bind(tx).Upsert(ctx, m); err != nil {
			return fmt.Errorf("upsert todo: %w", err)
		}
		return nil
	}); err != nil {
		return PutTodoOutput{}, fmt.Errorf("put todo: %w", err)
	}
	return PutTodoOutput{}, nil
}
