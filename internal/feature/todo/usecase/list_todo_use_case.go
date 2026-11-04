package usecase

import (
	"cmp"
	"context"
	"fmt"

	"github.com/go-kanna/showcase/internal/errors/aerrors"
	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/repository"
	"github.com/go-kanna/showcase/internal/feature/todo/scopes"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
	"github.com/go-kanna/showcase/internal/messages"
)

const (
	// defaultPageSize is the size of a page that a request leaves at 0.
	defaultPageSize = 20
	// latestComments is how many comments of each to-do a list reads.
	latestComments = 3
)

//kanna:validate
type ListTodoInput struct {
	Filter    ListFilter `json:"filter"`                              // nil lists every to-do
	PageSize  int32      `json:"page_size"  validate:"gte=0,lte=100"` // 0 for defaultPageSize
	PageToken string     `json:"page_token"`                          // empty for the first page
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
	Todos         []repository.TodoWithCommentCount
	NextPageToken string // empty on the last page
}

//kanna:container
type ListTodo struct {
	transactor tx.ReadTransactor `di:"arg"`
	todos      repository.Todo   `di:"arg"`
}

func (uc ListTodo) Do(ctx context.Context, in ListTodoInput) (ListTodoOutput, error) {
	// Before Validate, which a nil pointer to a variant would panic in: a pointer
	// implements ListFilter as well, and would otherwise list every to-do.
	s := scopes.Todos().CreatedAtAsc().
		// The limit keeps the latest comments of each to-do, not of the page.
		WithComments(scopes.Comments().CreatedAtDesc().Limit(latestComments))
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

	// The token holds the values of the last to-do of the page before, so the
	// page goes on from it however many to-dos were put or deleted meanwhile.
	s, err := s.AfterCursor(in.PageToken)
	if err != nil {
		return ListTodoOutput{}, aerrors.InvalidArgument("page_token").WithMessage(messages.TodoInvalidPageToken())
	}
	size := int(cmp.Or(in.PageSize, defaultPageSize))

	var ms []repository.TodoWithCommentCount
	if err := uc.transactor.WithReadTx(ctx, func(tx tx.Tx) error {
		// One more than the page holds tells whether a page follows.
		var err error
		ms, err = uc.todos.Bind(tx).ListWithCommentCounts(ctx, s.Limit(size+1))
		if err != nil {
			return fmt.Errorf("list todos: %w", err)
		}
		return nil
	}); err != nil {
		return ListTodoOutput{}, fmt.Errorf("list todos: %w", err)
	}
	if len(ms) <= size {
		return ListTodoOutput{Todos: ms}, nil
	}

	ms = ms[:size]
	next, err := s.Cursor(ms[size-1].Todo)
	if err != nil {
		return ListTodoOutput{}, fmt.Errorf("cursor: %w", err)
	}
	return ListTodoOutput{Todos: ms, NextPageToken: next}, nil
}
