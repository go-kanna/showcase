package usecase

import (
	"context"
	"fmt"
	"uuid"

	"github.com/go-kanna/showcase/internal/di"
	"github.com/go-kanna/showcase/internal/feature/todo/repository"
	"github.com/go-kanna/showcase/internal/feature/todo/scopes"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
	"github.com/go-kanna/showcase/internal/lib/either"
)

//kanna:validate
type DeleteTodoInput struct {
	TodoID string `json:"id" validate:"required,uuid"`
}

type DeleteTodoOutput struct {
}

//kanna:container
type DeleteTodo struct {
	_          *di.Infra `di:"arg"`
	transactor tx.Transactor
	todos      repository.Todo `di:"with=postgres.NewTodo"`
}

func (uc DeleteTodo) Do(ctx context.Context, in DeleteTodoInput) (DeleteTodoOutput, error) {
	if err := in.Validate(); err != nil {
		return DeleteTodoOutput{}, fmt.Errorf("validate input: %w", err)
	}

	id := either.Must(uuid.Parse(in.TodoID))
	if err := uc.transactor.WithTx(ctx, func(tx tx.Tx) error {
		if _, err := uc.todos.Bind(tx).Delete(ctx, scopes.Todos().IDEq(id)); err != nil {
			return fmt.Errorf("delete todo: %w", err)
		}
		return nil
	}); err != nil {
		return DeleteTodoOutput{}, fmt.Errorf("delete todo: %w", err)
	}
	return DeleteTodoOutput{}, nil
}
