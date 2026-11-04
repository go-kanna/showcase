package usecase_test

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/model/fixture"
	"github.com/go-kanna/showcase/internal/feature/todo/postgres"
	"github.com/go-kanna/showcase/internal/feature/todo/repository"
	"github.com/go-kanna/showcase/internal/feature/todo/usecase"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
	"github.com/go-kanna/showcase/test/infra"
)

// The latest comments come newest first, each with the to-do it is on, which
// the same query joins.
func TestListRecentComments(t *testing.T) {
	t.Parallel()

	// arrange
	ifr := infra.New(t)
	milk := fixture.Todo(func(m *model.Todo) { m.Title = "Buy milk" })
	dog := fixture.Todo(func(m *model.Todo) { m.Title = "Walk the dog" })
	var comments []model.Comment // as inserted, with the IDs the database gave
	ifr.InTx(t, func(tx tx.Tx) {
		for _, m := range []model.Todo{milk, dog} {
			require.NoError(t, postgres.NewTodo(ifr.Reader).Bind(tx).Upsert(t.Context(), m))
		}
		for _, todoID := range []uuid.UUID{milk.ID, dog.ID, milk.ID} {
			c, err := postgres.NewComment(ifr.Reader).Bind(tx).
				Insert(t.Context(), fixture.Comment(func(c *model.Comment) { c.TodoID = todoID }))
			require.NoError(t, err)
			comments = append(comments, c)
		}
	})
	sut := usecase.NewListRecentComments(ifr.ReadTransactor, postgres.NewComment(ifr.Reader))

	// act
	got, err := sut.Do(t.Context())

	// assert
	require.NoError(t, err)
	require.Len(t, got.Comments, 3)
	for i, want := range []struct {
		comment model.Comment
		todo    repository.TodoRef
	}{
		{comment: comments[2], todo: repository.TodoRef{ID: milk.ID, Title: "Buy milk"}},
		{comment: comments[1], todo: repository.TodoRef{ID: dog.ID, Title: "Walk the dog"}},
		{comment: comments[0], todo: repository.TodoRef{ID: milk.ID, Title: "Buy milk"}},
	} {
		assert.Equal(t, want.comment.ID, got.Comments[i].ID)
		assert.Equal(t, want.comment.Body, got.Comments[i].Body)
		assert.Equal(t, want.todo, got.Comments[i].Todo)
	}
}
