package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/go-kanna/validate"

	"github.com/go-kanna/showcase/internal/di"
	"github.com/go-kanna/showcase/internal/errors/aerrors"
	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/repository"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
	"github.com/go-kanna/showcase/internal/lib/either"
	"github.com/go-kanna/showcase/internal/lib/ptr"
)

//kanna:validate
type PutTodoInput struct {
	ID          string `validate:"required,uuid"`
	Title       string `validate:"required,min=1,max=200"`
	CompletedAt *time.Time
}

func (x PutTodoInput) validate(errs *validate.Errors) {
	if x.CompletedAt != nil {
		errs.Add("CompletedAt", model.CompletedAtRule{}.Validate(*x.CompletedAt))
	}
}

type PutTodoOutput struct {
}

//kanna:container
type PutTodo struct {
	_          *di.Infra `di:"arg"`
	transactor tx.Transactor
	todos      repository.Todo `di:"with=postgres.NewTodo"`
}

func (uc PutTodo) Do(ctx context.Context, in PutTodoInput) (PutTodoOutput, error) {
	if err := in.Validate(); err != nil {
		return PutTodoOutput{}, err
	}

	id := either.Must(uuid.Parse(in.ID))
	if err := uc.transactor.WithTx(ctx, func(tx tx.Tx) error {
		todos := uc.todos.Bind(tx)
		_, err := todos.Find(ctx, id)
		if err != nil && !errors.Is(err, aerrors.ErrNotFound) {
			return fmt.Errorf("find todo: %w", err)
		}

		now := time.Now()
		m := model.Todo{
			ID:          id,
			Title:       in.Title,
			CompletedAt: ptr.Map(in.CompletedAt, model.MustCompletedAt),
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		_, err = todos.Upsert(ctx, m)
		if err != nil {
			return fmt.Errorf("upsert todo: %w", err)
		}
		return nil
	}); err != nil {
		return PutTodoOutput{}, fmt.Errorf("put todo: %w", err)
	}
	return PutTodoOutput{}, nil
}
