package events

import (
	"context"
	"errors"
	"testing"
)

func Test_NoOpTracker(t *testing.T) {
	tracker := NewNoOpTracker()

	if _, err := tracker.Start(context.Background(), Event{}); err != nil {
		t.Error("noOpTracker must not retunr an error")
	}

	if _, err := tracker.End(context.Background(), Event{}, errors.New("test")); err != nil {
		t.Error("noOpTracker must not retunr an error")
	}
}
