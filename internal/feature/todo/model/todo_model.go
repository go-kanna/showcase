package model

import (
	"time"
	"uuid"

	"github.com/go-kanna/validated"
)

var (
	validCompletedAtStart = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	validCompletedAtEnd   = time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
)

//kanna:ignore
type CompletedAtRule struct{}

func (CompletedAtRule) Validate(v time.Time) error {
	switch {
	case v.Before(validCompletedAtStart):
		return validated.FailWith("greater_than_or_equal_to", map[string]any{"count": validCompletedAtStart})
	case v.After(validCompletedAtEnd):
		return validated.FailWith("less_than_or_equal_to", map[string]any{"count": validCompletedAtEnd})
	}
	return nil
}

type CompletedAt = validated.Value[time.Time, CompletedAtRule]

func MustCompletedAt(v time.Time) CompletedAt {
	return validated.Must[CompletedAtRule](v)
}

//go:generate go tool kanna-fixture

type Todo struct {
	ID          uuid.UUID
	Title       string
	CompletedAt *CompletedAt
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
