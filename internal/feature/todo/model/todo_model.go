package model

import (
	"time"
	"uuid"

	"github.com/go-kanna/value/rule"
)

var (
	validCompletedAtStart = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	validCompletedAtEnd   = time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
)

//go:generate go tool kanna-value

//kanna:value
type CompletedAt struct{ t time.Time }

func (CompletedAt) validate(t time.Time) error {
	switch {
	case t.Before(validCompletedAtStart):
		return rule.FailWith("greater_than_or_equal_to", map[string]any{"count": validCompletedAtStart})
	case t.After(validCompletedAtEnd):
		return rule.FailWith("less_than_or_equal_to", map[string]any{"count": validCompletedAtEnd})
	}
	return nil
}

// Priority is how soon a to-do is to be done. Only its constants come in from
// a request or a row.
//
//kanna:enum
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
)

//go:generate go tool kanna-fixture

type Todo struct {
	ID          uuid.UUID
	Title       string
	Priority    Priority
	CompletedAt *CompletedAt
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
