package events

import (
	"context"
)

// Tracker - An interface for tracking events
type Tracker interface {
	Start(context.Context, Event) (context.Context, error)
	End(context.Context, Event, error) (context.Context, error)
}

// NewNoOpTracker - Returns a "No Operation Tracker"
func NewNoOpTracker() Tracker {
	return &noOpTracker{}
}

type noOpTracker struct{}

func (t *noOpTracker) Start(ctx context.Context, event Event) (context.Context, error) {
	return ctx, nil
}

func (t *noOpTracker) End(ctx context.Context, event Event, err error) (context.Context, error) {
	return ctx, nil
}
