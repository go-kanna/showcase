package mapper

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/go-kanna/showcase/internal/feature/todo/model"
	todov1 "github.com/go-kanna/showcase/internal/gen/showcase/todo/v1"
	"github.com/go-kanna/showcase/internal/lib/converter"
)

//go:generate go tool kanna-mapper

//kanna:mapper
type _ interface {
	TodoToProto(m model.Todo) *todov1.Todo
}

//kanna:converter
var _ = []any{
	converter.Defaults,
	mapCompletedAt,
}

func mapCompletedAt(at *model.CompletedAt) *timestamppb.Timestamp {
	if at == nil {
		return nil
	}
	return timestamppb.New(at.Get())
}
