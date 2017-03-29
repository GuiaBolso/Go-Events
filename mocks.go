package events

import (
	"context"
	"net/http"
)

//MockTracker - A genereic mock for Tracker interface
type MockTracker struct {
	StartFn            func(context.Context, Event, http.ResponseWriter, *http.Request) context.Context
	NoticeErrorFn      func(context.Context, error) context.Context
	NoticeEventErrorFn func(context.Context, Event, error) context.Context
	EndFn              func(context.Context, Event, error) context.Context

	StartCount            int
	NoticeErrorCount      int
	NoticeEventErrorCount int
	EndCount              int
}

// Start - Imcrements the counter and calls the mock implementation
func (t *MockTracker) Start(ctx context.Context, event Event, w http.ResponseWriter, r *http.Request) context.Context {
	t.StartCount++
	return t.StartFn(ctx, event, w, r)
}

// NoticeError - Imcrements the counter and calls the mock implementation
func (t *MockTracker) NoticeError(ctx context.Context, err error) context.Context {
	t.NoticeErrorCount++
	return t.NoticeErrorFn(ctx, err)
}

// NoticeEventError - Imcrements the counter and calls the mock implementation
func (t *MockTracker) NoticeEventError(ctx context.Context, event Event, err error) context.Context {
	t.NoticeEventErrorCount++
	return t.NoticeEventErrorFn(ctx, event, err)
}

//End - Imcrements the counter and calls the mock implementation
func (t *MockTracker) End(ctx context.Context, event Event, err error) context.Context {
	t.EndCount++
	return t.EndFn(ctx, event, err)
}
