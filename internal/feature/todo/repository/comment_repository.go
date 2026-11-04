package repository

import (
	"context"

	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
)

//kanna:repository
type Comment interface {
	Insert(ctx context.Context, c model.Comment) (model.Comment, error)
	Bind(tx tx.Tx) Comment
}
