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

//kanna:validate
type ListTodoInput struct {
	Filter ListFilter `json:"filter"` // nil lists every to-do
}

// ListFilter narrows a list to one condition. It is a sum type: one of
// CompletionFilter and PriorityFilter, which kanna-mapper matches to the arms
// of the request's oneof by name.
type ListFilter interface {
	Validate() error
	isListFilter()
}

// CompletionFilter lists the to-dos completed, or those not.
type CompletionFilter struct {
	Completed bool `json:"completed"`
}

// PriorityFilter lists the to-dos of one priority.
//
//kanna:validate
type PriorityFilter struct {
	Priority model.Priority `json:"priority" validate:"required"`
}

// Validate accepts any CompletionFilter: both answers are a filter.
func (CompletionFilter) Validate() error { return nil }

func (CompletionFilter) isListFilter() {}
func (PriorityFilter) isListFilter()   {}

type ListTodoOutput struct {
	Todos []model.Todo
}

//kanna:container
type ListTodo struct {
	transactor tx.ReadTransactor `di:"arg"`
	todos      repository.Todo   `di:"arg"`
}

func (uc ListTodo) Do(ctx context.Context, in ListTodoInput) (ListTodoOutput, error) {
	if err := in.Validate(); err != nil {
		return ListTodoOutput{}, err
	}

	s := scopes.Todos().CreatedAtAsc()
	switch f := in.Filter.(type) {
	case CompletionFilter:
		if f.Completed {
			s = s.CompletedAtIsNotNull()
		} else {
			s = s.CompletedAtIsNull()
		}
	case PriorityFilter:
		s = s.PriorityEq(f.Priority)
	}

	var ms []model.Todo
	if err := uc.transactor.WithReadTx(ctx, func(tx tx.Tx) error {
		todos := uc.todos.Bind(tx)
		var err error
		ms, err = todos.List(ctx, s)
		if err != nil && !errors.Is(err, aerrors.ErrNotFound) {
			return fmt.Errorf("find todo: %w", err)
		}
		return nil
	}); err != nil {
		return ListTodoOutput{}, fmt.Errorf("put todo: %w", err)
	}
	return ListTodoOutput{Todos: ms}, nil
}
