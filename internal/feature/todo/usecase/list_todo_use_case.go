package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kanna/showcase/internal/errors/aerrors"
	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/repository"
	"github.com/go-kanna/showcase/internal/feature/todo/scopes"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
)

type ListTodoInput struct {
}

type ListTodoOutput struct {
	Todos []model.Todo
}

//kanna:container
type ListTodo struct {
	transactor tx.ReadTransactor `di:"arg"`
	todos      repository.Todo   `di:"arg"`
}

func (uc ListTodo) Do(ctx context.Context, in ListTodoInput) (ListTodoOutput, error) {
	var ms []model.Todo
	if err := uc.transactor.WithReadTx(ctx, func(tx tx.Tx) error {
		todos := uc.todos.Bind(tx)
		var err error
		ms, err = todos.List(ctx, scopes.Todos().CreatedAtAsc())
		if err != nil && !errors.Is(err, aerrors.ErrNotFound) {
			return fmt.Errorf("find todo: %w", err)
		}
		return nil
	}); err != nil {
		return ListTodoOutput{}, fmt.Errorf("put todo: %w", err)
	}
	return ListTodoOutput{Todos: ms}, nil
}
