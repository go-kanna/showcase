package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/go-kanna/showcase/internal/errors/aerrors"
	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/repository"
	"github.com/go-kanna/showcase/internal/feature/todo/scopes"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
	"github.com/go-kanna/showcase/internal/lib/either"
	"github.com/go-kanna/showcase/internal/messages"
)

//kanna:validate
type SetTodoCompletionInput struct {
	TodoID    string `json:"id" validate:"required,uuid"`
	Completed bool
}

type SetTodoCompletionOutput struct {
}

//kanna:container
type SetTodoCompletion struct {
	transactor tx.Transactor   `di:"arg"`
	todos      repository.Todo `di:"arg"`
}

func (uc SetTodoCompletion) Do(ctx context.Context, in SetTodoCompletionInput) (SetTodoCompletionOutput, error) {
	if err := in.Validate(); err != nil {
		return SetTodoCompletionOutput{}, err
	}

	id := either.Must(uuid.Parse(in.TodoID))
	if err := uc.transactor.WithTx(ctx, func(tx tx.Tx) error {
		todos := uc.todos.Bind(tx)
		m, err := todos.GetForUpdate(ctx, id)
		if errors.Is(err, aerrors.ErrNotFound) {
			return aerrors.NotFound("todo").WithMessage(messages.TodoNotFound())
		}
		if err != nil {
			return fmt.Errorf("find todo: %w", err)
		}
		// A to-do already in the state asked for is left as it is, its
		// updated_at included.
		if in.Completed == (m.CompletedAt != nil) {
			return nil
		}

		now := time.Now()
		c := scopes.TodoChanges()
		if in.Completed {
			at, err := model.NewCompletedAt(now)
			if err != nil {
				return fmt.Errorf("completed at: %w", err)
			}
			c = c.SetCompletedAt(&at)
		} else {
			c = c.SetCompletedAt(nil)
		}
		if err := todos.Update(ctx, id, c); err != nil {
			return fmt.Errorf("update todo: %w", err)
		}
		return nil
	}); err != nil {
		return SetTodoCompletionOutput{}, fmt.Errorf("set todo completion: %w", err)
	}
	return SetTodoCompletionOutput{}, nil
}
