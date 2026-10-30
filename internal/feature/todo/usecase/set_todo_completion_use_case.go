package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/go-kanna/orm"

	"github.com/go-kanna/showcase/internal/errors/aerrors"
	"github.com/go-kanna/showcase/internal/feature/todo/repository"
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
		_, err := todos.Get(ctx, id)
		if errors.Is(err, orm.ErrNotFound) {
			return aerrors.NotFound("todo").WithMessage(messages.TodoNotFound())
		}
		if err != nil {
			return fmt.Errorf("find todo: %w", err)
		}
		if in.Completed {
			_, err := todos.Complete(ctx, id, time.Now())
			if err != nil {
				return fmt.Errorf("complete: %w", err)
			}
			return nil
		}
		_, err = todos.Uncomplete(ctx, id)
		if err != nil {
			return fmt.Errorf("uncomplete: %w", err)
		}
		return nil
	}); err != nil {
		return SetTodoCompletionOutput{}, fmt.Errorf("set todo completion: %w", err)
	}
	return SetTodoCompletionOutput{}, nil
}
