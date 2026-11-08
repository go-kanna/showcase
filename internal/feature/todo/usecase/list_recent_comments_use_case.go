package usecase

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/go-kanna/showcase/internal/feature/todo/repository"
	"github.com/go-kanna/showcase/internal/feature/todo/scopes"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
	"github.com/go-kanna/showcase/internal/lib/xslices"
)

// recentComments is how many comments a list of the recent ones holds.
const recentComments = 20

type ListRecentCommentsOutput struct {
	Comments []CommentSummary
}

// CommentSummary is a comment with the to-do it is on.
type CommentSummary struct {
	ID        uuid.UUID
	Body      string
	CreatedAt time.Time
	Todo      TodoRef
}

// TodoRef names a to-do.
type TodoRef struct {
	ID    uuid.UUID
	Title string
}

//kanna:container
type ListRecentComments struct {
	transactor tx.ReadTransactor  `di:"arg"`
	comments   repository.Comment `di:"arg"`
}

func (uc ListRecentComments) Do(ctx context.Context) (ListRecentCommentsOutput, error) {
	var cs []repository.CommentSummary
	if err := uc.transactor.WithReadTx(ctx, func(tx tx.Tx) error {
		var err error
		cs, err = uc.comments.Bind(tx).ListSummaries(ctx, scopes.Comments().CreatedAtDesc().Limit(recentComments))
		if err != nil {
			return fmt.Errorf("list comments: %w", err)
		}
		return nil
	}); err != nil {
		return ListRecentCommentsOutput{}, fmt.Errorf("list recent comments: %w", err)
	}
	return ListRecentCommentsOutput{Comments: xslices.Map(cs, commentSummaryFromRepository)}, nil
}
