package infra

import (
	"context"
	"time"
)

// Clock tells the time it holds, for a test to say what the repositories write
// as the times rows are created and updated, through option.WithClock.
type Clock struct{ T time.Time }

func (c *Clock) Now(context.Context) time.Time { return c.T }
