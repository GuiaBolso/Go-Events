package events

import (
	"context"
	"net/http"
)

// HTTPTracker - An interface for tracking events
type HTTPTracker interface {
	Start(context.Context, Event, http.ResponseWriter, *http.Request) context.Context
	End(context.Context, Event, error) context.Context
}

// NewNoOpTracker - Returns a "No Operation Tracker"
func NewNoOpTracker() HTTPTracker {
	return &noOpTracker{}
}

type noOpTracker struct{}

func (t *noOpTracker) Start(ctx context.Context, event Event, w http.ResponseWriter, r *http.Request) context.Context {
	return ctx
}

func (t *noOpTracker) End(ctx context.Context, event Event, err error) context.Context {
	return ctx
}
