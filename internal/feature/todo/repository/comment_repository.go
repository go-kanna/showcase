package repository

import (
	"context"

	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/projection"
	"github.com/go-kanna/showcase/internal/feature/todo/scopes"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
)

//kanna:repository
type Comment interface {
	Insert(ctx context.Context, c model.Comment) (model.Comment, error)
	ListSummaries(ctx context.Context, s ...scopes.Comment) ([]projection.CommentSummary, error)
	Bind(tx tx.Tx) Comment
}
