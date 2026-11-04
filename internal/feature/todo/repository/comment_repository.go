package repository

import (
	"context"
	"time"
	"uuid"

	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/scopes"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
)

//kanna:repository
type Comment interface {
	Insert(ctx context.Context, c model.Comment) (model.Comment, error)
	ListSummaries(ctx context.Context, s ...scopes.Comment) ([]CommentSummary, error)
	Bind(tx tx.Tx) Comment
}

// CommentSummary is a comment with the to-do it is on. kanna-orm reads Todo
// from todos by comments.todo_id, the field's name and _id, in an INNER JOIN,
// as Todo is a struct and not a pointer.
type CommentSummary struct {
	ID        uuid.UUID
	Body      string
	CreatedAt time.Time
	Todo      TodoRef
}

// TodoRef names a to-do.
type TodoRef struct {
	ID    uuid.UUID
	Title string
}
