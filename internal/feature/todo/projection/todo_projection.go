// Package projection holds what the queries of the repositories return besides
// the models, for every layer to read.
package projection

import (
	"time"
	"uuid"

	"github.com/go-kanna/showcase/internal/feature/todo/model"
)

// TodoWithCommentCount is a to-do with how many comments it has, which the
// query adds to its row. The count stays out of model.Todo, which every other
// method reads without it.
type TodoWithCommentCount struct {
	model.Todo

	CommentCount int64
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
