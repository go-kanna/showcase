package usecase

import (
	"context"
	"fmt"

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
// Completion and Priority, which kanna-mapper matches to the arms of the
// request's oneof by name.
type ListFilter interface {
	Validate() error
	isListFilter()
}

// Completion lists the to-dos completed, or those not.
type Completion struct {
	Completed bool `json:"completed"`
}

// Priority lists the to-dos of one priority.
//
//kanna:validate
type Priority struct {
	Priority model.Priority `json:"priority" validate:"required"`
}

// Validate accepts any Completion: both answers are a filter.
func (Completion) Validate() error { return nil }

func (Completion) isListFilter() {}
func (Priority) isListFilter()   {}

type ListTodoOutput struct {
	Todos []model.Todo
}

//kanna:container
type ListTodo struct {
	transactor tx.ReadTransactor `di:"arg"`
	todos      repository.Todo   `di:"arg"`
}

func (uc ListTodo) Do(ctx context.Context, in ListTodoInput) (ListTodoOutput, error) {
	// Before Validate, which a nil pointer to a variant would panic in: a pointer
	// implements ListFilter as well, and would otherwise list every to-do.
	s := scopes.Todos().CreatedAtAsc()
	switch f := in.Filter.(type) {
	case nil:
	case Completion:
		if f.Completed {
			s = s.CompletedAtIsNotNull()
		} else {
			s = s.CompletedAtIsNull()
		}
	case Priority:
		s = s.PriorityEq(f.Priority)
	default:
		return ListTodoOutput{}, fmt.Errorf("unsupported filter %T", f)
	}

	if err := in.Validate(); err != nil {
		return ListTodoOutput{}, err
	}

	var ms []model.Todo
	if err := uc.transactor.WithReadTx(ctx, func(tx tx.Tx) error {
		todos := uc.todos.Bind(tx)
		var err error
		ms, err = todos.List(ctx, s)
		if err != nil {
			return fmt.Errorf("list todos: %w", err)
		}
		return nil
	}); err != nil {
		return ListTodoOutput{}, fmt.Errorf("list todos: %w", err)
	}
	return ListTodoOutput{Todos: ms}, nil
}
