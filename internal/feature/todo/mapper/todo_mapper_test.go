package mapper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-kanna/showcase/internal/feature/todo/mapper"
	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/model/fixture"
	todov1 "github.com/go-kanna/showcase/internal/gen/showcase/todo/v1"
)

func TestTodoToProto_Priority(t *testing.T) {
	t.Parallel()

	// arrange
	m := fixture.Todo(func(m *model.Todo) { m.Priority = model.PriorityHigh })

	// act
	got := mapper.TodoToProto(m)

	// assert
	assert.Equal(t, todov1.Priority_PRIORITY_HIGH, got.GetPriority())
}

func TestTodoInputFromProto_Priority(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      todov1.Priority
		want    model.Priority
		wantErr string
	}{
		{name: "a value the model has", in: todov1.Priority_PRIORITY_LOW, want: model.PriorityLow},
		// validate's required reports the zero value as blank.
		{name: "left unspecified", in: todov1.Priority_PRIORITY_UNSPECIFIED, want: ""},
		{
			name:    "a number nobody declared",
			in:      todov1.Priority(9),
			wantErr: "map usecase.PutTodoInput.Priority: model.Priority has no value for 9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// act
			got, err := mapper.TodoInputFromProto(&todov1.TodoInput{Id: "id", Title: "title", Priority: tt.in})

			// assert
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Priority)
		})
	}
}
