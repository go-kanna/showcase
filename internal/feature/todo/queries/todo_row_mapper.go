package queries

import (
	"time"

	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/lib/converter"
)

//go:generate go tool kanna-mapper

//kanna:mapper
type _ interface {
	RowToTodo(m Todo) model.Todo
}

//kanna:converter
var _ = []any{
	converter.Defaults,
	mapCompletedAt,
}

func mapCompletedAt(at *time.Time) *model.CompletedAt {
	if at == nil {
		return nil
	}
	return new(model.MustCompletedAt(*at))
}
