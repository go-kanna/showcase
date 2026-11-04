package postgres

import (
	"github.com/go-kanna/showcase/internal/feature/todo/repository"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
)

func (r Comment) Bind(tx tx.Tx) repository.Comment {
	return NewComment(tx.DBTX())
}
