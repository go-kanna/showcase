package handler_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/go-kanna/showcase/internal/errors/aerrors"
	"github.com/go-kanna/showcase/internal/feature/todo/handler"
	todov1 "github.com/go-kanna/showcase/internal/gen/showcase/todo/v1"
)

// A timestamp protobuf does not allow is rejected before it reaches the use
// case, which the zero Handler does not have: reaching it would panic.
func TestHandler_PutTodo_InvalidCompletedAt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ts   *timestamppb.Timestamp
	}{
		{name: "negative nanos", ts: &timestamppb.Timestamp{Seconds: 1767225600, Nanos: -1}},
		{name: "a second of nanos", ts: &timestamppb.Timestamp{Seconds: 1767225600, Nanos: 1_000_000_000}},
		{name: "past the year 9999", ts: &timestamppb.Timestamp{Seconds: math.MaxInt64}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// act
			_, err := handler.Handler{}.PutTodo(t.Context(), &todov1.PutTodoRequest{
				Todo:        &todov1.TodoInput{Id: "01890a5d-ac96-774b-bcce-b302099a8057", Title: "Buy milk"},
				CompletedAt: tt.ts,
			})

			// assert
			require.Error(t, err)
			assert.ErrorIs(t, err, aerrors.ErrInvalidArgument)
		})
	}
}
