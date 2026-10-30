package mapper

import (
	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/usecase"
	todov1 "github.com/go-kanna/showcase/internal/gen/showcase/todo/v1"
	"github.com/go-kanna/showcase/internal/lib/converter"
)

//go:generate go tool kanna-mapper

//kanna:mapper
type _ interface {
	TodoToProto(m model.Todo) *todov1.Todo

	// TodoInputFromProto reads the to-do a PutTodo request carries. The
	// request carries its completion apart from it.
	//
	//kanna:map CompletedAt=-
	TodoInputFromProto(in *todov1.TodoInput) (usecase.PutTodoInput, error)
}

//kanna:converter
var _ = []any{
	converter.Defaults,
}
