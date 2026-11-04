package postgres

import (
	"errors"

	"github.com/go-kanna/orm"

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
