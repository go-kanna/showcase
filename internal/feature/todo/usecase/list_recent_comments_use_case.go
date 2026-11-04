package usecase

import (
	"context"
	"fmt"

	"github.com/go-kanna/showcase/internal/feature/todo/repository"
	"github.com/go-kanna/showcase/internal/feature/todo/scopes"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
)

// recentComments is how many comments a list of the recent ones holds.
const recentComments = 20

type ListRecentCommentsOutput struct {
	Comments []repository.CommentSummary
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
	return ListRecentCommentsOutput{Comments: cs}, nil
}
