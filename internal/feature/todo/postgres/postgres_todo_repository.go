package postgres

import (
	"github.com/go-kanna/showcase/internal/feature/todo/repository"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
)

func (r Todo) Bind(tx tx.Tx) repository.Todo {
	return NewTodo(tx.DBTX())
}
