package server_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-kanna/validate"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/go-kanna/showcase/internal/errors/aerrors"
	"github.com/go-kanna/showcase/internal/messages"
	"github.com/go-kanna/showcase/internal/server"
)

func TestToStatus(t *testing.T) {
	t.Parallel()

	var verrs validate.Errors
	verrs.Add("id", validate.Blank())
	verrs.Add("title", validate.TooLong(200))
	notFound := aerrors.NotFound("todo").WithMessage(messages.TodoNotFound())

	tests := []struct {
		name     string
		lang     string
		err      error
		wantCode codes.Code
		wantMsg  string
	}{
		{
			name:     "validation errors, in the default language",
			err:      verrs.Err(),
			wantCode: codes.InvalidArgument,
			wantMsg:  "id can't be blank; title is too long (maximum is 200 characters)",
		},
		{
			name:     "validation errors, in the requested language",
			lang:     "ja",
			err:      verrs.Err(),
			wantCode: codes.InvalidArgument,
			wantMsg:  "id を入力してください; title は200文字以内で入力してください", //nolint:gosmopolitan // Japanese is under test
		},
		{
			name:     "wrapped app error without a message",
			err:      fmt.Errorf("delete todo: %w", aerrors.Conflict("todo")),
			wantCode: codes.AlreadyExists,
			wantMsg:  "todo: already exists",
		},
		{
			name:     "app error with a message, in the default language",
			err:      fmt.Errorf("set todo completion: %w", notFound),
			wantCode: codes.NotFound,
			wantMsg:  "To-do not found.",
		},
		{
			name:     "app error with a message, in the requested language",
			lang:     "ja",
			err:      notFound,
			wantCode: codes.NotFound,
			wantMsg:  "やることが見つかりませんでした。", //nolint:gosmopolitan // Japanese is under test
		},
		{
			name:     "canceled",
			err:      fmt.Errorf("list todo: %w", context.Canceled),
			wantCode: codes.Canceled,
		},
		{
			name:     "anything else",
			err:      errors.New("connection refused"),
			wantCode: codes.Internal,
			wantMsg:  "internal error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			ctx := t.Context()
			if tt.lang != "" {
				ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("accept-language", tt.lang))
			}

			// act
			got := status.Convert(server.ToStatus(ctx, "/showcase.todo.v1.TodoService/PutTodo", tt.err))

			// assert
			assert.Equal(t, tt.wantCode, got.Code())
			if tt.wantMsg != "" {
				assert.Equal(t, tt.wantMsg, got.Message())
			}
		})
	}
}
