package aerrors

import (
	"errors"

	"github.com/go-kanna/i18n"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("already exists")
	ErrPrecondition = errors.New("precondition failed")

	ErrUnauthenticated   = errors.New("unauthenticated")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrResourceExhausted = errors.New("resource exhausted")
	ErrAborted           = errors.New("aborted")
)

type Error struct {
	detail  string
	kind    error
	message i18n.Message
}

func (e Error) Error() string {
	return e.detail + ": " + e.kind.Error()
}

func (e Error) Unwrap() error {
	return e.kind
}

func (e Error) WithMessage(m i18n.Message) Error {
	e.message = m
	return e
}

func (e Error) Message() (i18n.Message, bool) {
	return e.message, e.message.Key != ""
}

func NotFound(entity string) Error {
	return Error{detail: entity, kind: ErrNotFound}
}

func Conflict(detail string) Error {
	return Error{detail: detail, kind: ErrConflict}
}

func Precondition(detail string) Error {
	return Error{detail: detail, kind: ErrPrecondition}
}

func Unauthenticated(detail string) Error {
	return Error{detail: detail, kind: ErrUnauthenticated}
}

func PermissionDenied(detail string) Error {
	return Error{detail: detail, kind: ErrPermissionDenied}
}

func InvalidArgument(detail string) Error {
	return Error{detail: detail, kind: ErrInvalidArgument}
}

func ResourceExhausted(detail string) Error {
	return Error{detail: detail, kind: ErrResourceExhausted}
}

func Aborted(detail string) Error {
	return Error{detail: detail, kind: ErrAborted}
}
