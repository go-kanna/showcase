package postgres

import (
	"github.com/go-kanna/showcase/internal/feature/todo/repository"
	"github.com/go-kanna/showcase/internal/infra/storage/db"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
)

// NewDBTX gives the repositories the reader's pool to run on outside a transaction; Bind moves one into a
// transaction to write.
func NewDBTX(reader db.Reader) DBTX {
	return reader
}

func (r Todo) Bind(tx tx.Tx) repository.Todo {
	return NewTodo(tx.DBTX())
}
