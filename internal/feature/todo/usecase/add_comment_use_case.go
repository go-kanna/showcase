package usecase

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/go-kanna/showcase/internal/errors/aerrors"
	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/repository"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
	"github.com/go-kanna/showcase/internal/lib/either"
	"github.com/go-kanna/showcase/internal/messages"
)

//kanna:validate
type AddCommentInput struct {
	TodoID string `json:"todo_id" validate:"required,uuid"`
	Body   string `json:"body"    validate:"required,max=1000"`
}

type AddCommentOutput struct {
	Comment model.Comment
}

//kanna:container
type AddComment struct {
	transactor tx.Transactor      `di:"arg"`
	comments   repository.Comment `di:"arg"`
}

func (uc AddComment) Do(ctx context.Context, in AddCommentInput) (AddCommentOutput, error) {
	if err := in.Validate(); err != nil {
		return AddCommentOutput{}, err
	}

	todoID := either.Must(uuid.Parse(in.TodoID))
	var c model.Comment
	if err := uc.transactor.WithTx(ctx, func(tx tx.Tx) error {
		// The database gives the comment its ID, which Insert reads back, and
		// refuses a comment on a to-do that is not there.
		var err error
		c, err = uc.comments.Bind(tx).Insert(ctx, model.Comment{TodoID: todoID, Body: in.Body})
		if errors.Is(err, aerrors.ErrNotFound) {
			return aerrors.NotFound("todo").WithMessage(messages.TodoNotFound())
		}
		if err != nil {
			return fmt.Errorf("insert comment: %w", err)
		}
		return nil
	}); err != nil {
		return AddCommentOutput{}, fmt.Errorf("add comment: %w", err)
	}
	return AddCommentOutput{Comment: c}, nil
}
