package usecase_test

import (
	"strings"
	"testing"
	"uuid"

	"github.com/go-kanna/validate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-kanna/showcase/internal/errors/aerrors"
	"github.com/go-kanna/showcase/internal/feature/todo/model/fixture"
	"github.com/go-kanna/showcase/internal/feature/todo/postgres"
	"github.com/go-kanna/showcase/internal/feature/todo/usecase"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
	"github.com/go-kanna/showcase/test/infra"
)

func TestAddComment(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	m := fixture.Todo()
	ifr.InTx(t, func(tx tx.Tx) {
		require.NoError(t, postgres.NewTodo(ifr.Reader).Bind(tx).Upsert(t.Context(), m))
	})
	sut := usecase.NewAddComment(ifr.Transactor, postgres.NewTodo(ifr.Reader), postgres.NewComment(ifr.Reader))

	// act
	got, err := sut.Do(t.Context(), usecase.AddCommentInput{TodoID: m.ID.String(), Body: "Whole milk"})

	// assert
	require.NoError(t, err)
	assert.NotEqual(t, uuid.UUID{}, got.Comment.ID, "the database gives the ID")
	assert.Equal(t, m.ID, got.Comment.TodoID)
	assert.Equal(t, "Whole milk", got.Comment.Body)
	listed, err := usecase.NewListTodo(ifr.ReadTransactor, postgres.NewTodo(ifr.Reader)).
		Do(t.Context(), usecase.ListTodoInput{})
	require.NoError(t, err)
	require.Len(t, listed.Todos, 1)
	require.Len(t, listed.Todos[0].Comments, 1)
	assert.Equal(t, got.Comment.ID, listed.Todos[0].Comments[0].ID)
}

func TestAddComment_TodoNotFound(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	sut := usecase.NewAddComment(ifr.Transactor, postgres.NewTodo(ifr.Reader), postgres.NewComment(ifr.Reader))

	// act
	_, err := sut.Do(t.Context(), usecase.AddCommentInput{TodoID: fixture.Todo().ID.String(), Body: "Whole milk"})

	// assert
	require.ErrorIs(t, err, aerrors.ErrNotFound)
}

// An input that breaks a rule is reported field by field, before any
// repository is called.
func TestAddComment_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		in    usecase.AddCommentInput
		field string
		code  string
	}{
		{
			name:  "no to-do",
			in:    usecase.AddCommentInput{Body: "Whole milk"},
			field: "todo_id",
			code:  "blank",
		},
		{
			name:  "no body",
			in:    usecase.AddCommentInput{TodoID: fixture.Todo().ID.String()},
			field: "body",
			code:  "blank",
		},
		{
			name:  "a body too long",
			in:    usecase.AddCommentInput{TodoID: fixture.Todo().ID.String(), Body: strings.Repeat("a", 1001)},
			field: "body",
			code:  "too_long",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// act
			_, err := usecase.NewAddComment(nil, nil, nil).Do(t.Context(), tt.in)

			// assert
			var verrs validate.Errors
			require.ErrorAs(t, err, &verrs)
			require.Len(t, verrs, 1)
			assert.Equal(t, tt.field, verrs[0].Field)
			assert.Equal(t, tt.code, verrs[0].Code())
		})
	}
}
