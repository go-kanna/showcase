package postgres

import (
	"errors"

	"github.com/go-kanna/orm"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/go-kanna/showcase/internal/errors/aerrors"
)

// mapError turns the errors of the database into the application's, which the
// generated methods return, so that the use cases see none of the database's.
func (Todo) mapError(err error) error {
	if errors.Is(err, orm.ErrNotFound) {
		return aerrors.NotFound("todo")
	}
	return err
}

// mapError says that the to-do a comment Insert writes is not there, as the
// foreign key of todo_id refuses it, so that AddComment needs not look it up
// first.
func (Comment) mapError(op commentOperation, err error) error {
	pgerr, ok := errors.AsType[*pgconn.PgError](err)
	if ok && op == commentOperationInsert && pgerr.Code == "23503" { // foreign_key_violation
		return aerrors.NotFound("todo")
	}
	return err
}
