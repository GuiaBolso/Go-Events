package events

import (
	"context"
	"encoding/json"
	"fmt"
)

// Event implementation
type Event struct {
	Name     string          `json:"name"`
	Version  int             `json:"version"`
	ID       string          `json:"id"`
	FlowID   string          `json:"flowId,omitempty"`
	Payload  json.RawMessage `json:"payload"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

type errorPayload struct {
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}

// NewError returns a error event
func NewError(flowID, message string) Event {
	payload, _ := json.Marshal(errorPayload{
		Message: message,
	})

	return Event{
		Name:    "error",
		Version: 1,
		ID:      RandomID(),
		FlowID:  flowID,
		Payload: payload,
	}
}

// NewErrorWithMetadata - Returns an error event with metadata
func NewErrorWithMetadata(flowID, message string, metadata interface{}) Event {
	payload, _ := json.Marshal(errorPayload{
		Message: message,
	})

	metadataJSON, _ := json.Marshal(metadata)

	return Event{
		Name:     "error",
		Version:  1,
		ID:       RandomID(),
		FlowID:   flowID,
		Payload:  payload,
		Metadata: metadataJSON,
	}
}

// NewResponse builds a response event
func NewResponse(request Event, payload interface{}) (Event, error) {
	response, err := json.Marshal(payload)

	return Event{
		Name:    fmt.Sprintf("%s:response", request.Name),
		Version: request.Version,
		ID:      RandomID(),
		FlowID:  request.FlowID,
		Payload: response,
	}, err
}

// Batch is an event that executes batches of events
func Batch(mux *Mux) HandlerFunc {
	return func(ctx context.Context, event Event) (Event, error) {
		payload := struct {
			Parallel bool    `json:"parallel"`
			Events   []Event `json:"events"`
		}{}

		responses := []Event{}

		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return NewError(event.FlowID, err.Error()), err
		}

		for _, ev := range payload.Events {
			if h, ok := mux.get(ev.Name, ev.Version); ok {
				resp, err := h.Serve(ctx, ev)

				responses = append(responses, resp)

				if err != nil {
					ctx = mux.tracer.NoticeEventError(ctx, ev, err)
				}

			} else {
				responses = append(responses, NewError(
					ev.FlowID,
					fmt.Sprintf(`Event "%s" not found`, ev.Name),
				))
			}
		}

		return NewResponse(event, responses)
	}
}
