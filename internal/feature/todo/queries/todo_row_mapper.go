package queries

import (
	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/lib/converter"
)

//go:generate go tool kanna-mapper

//kanna:mapper
type _ interface {
	RowToTodo(m Todo) (model.Todo, error)
}

//kanna:converter
var _ = []any{
	converter.Defaults,
}
