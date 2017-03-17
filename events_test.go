package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_Batch(t *testing.T) {
	mux := NewMux()
	mockEvents := []Event{}

	for i := 0; i < 4; i++ {
		mockEvents = append(mockEvents, Event{
			Name:    fmt.Sprintf("Mock-%d", i),
			Version: i + 1,
			ID:      RandomID(),
			FlowID:  RandomID(),
		})
	}

	mux.Add("Mock-0", 1, HandlerFunc(mockHandlerFunc))
	mux.Add("Mock-1", 2, HandlerFunc(mockHandlerFunc))
	mux.Add("Mock-2", 3, HandlerFunc(mockEventError))

	mux.Add("batch", 1, Batch(mux))

	batchPayload := struct {
		Parallel bool    `json:"parallel"`
		Events   []Event `json:"events"`
	}{
		Parallel: false,
		Events:   mockEvents,
	}

	batchPayloadJSON, _ := json.Marshal(batchPayload)

	event := Event{
		Name:    "batch",
		Version: 1,
		ID:      RandomID(),
		FlowID:  RandomID(),
		Payload: json.RawMessage(batchPayloadJSON),
	}

	eventJSON, _ := json.Marshal(event)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/events/", bytes.NewReader(eventJSON))

	mux.ServeHTTP(w, r)

	responseJSON := Event{}
	json.NewDecoder(w.Body).Decode(&responseJSON)

	responseEvents := []Event{}
	json.Unmarshal(responseJSON.Payload, &responseEvents)

	if responseJSON.FlowID != event.FlowID {
		t.Error("Expecting the same flowID as the request event")
	}

	if responseJSON.Name != "batch:response" {
		t.Errorf("Expecting a different response. %s", string(responseJSON.Payload))
	}

	if len(responseEvents) != 4 {
		t.Errorf("len(responses) == %d, wants: %d", len(responseEvents), 4)
	}
}

func Test_Batch_input_error(t *testing.T) {
	event := Event{
		Name:    "batch",
		Version: 1,
		ID:      RandomID(),
		FlowID:  RandomID(),
		Payload: json.RawMessage([]byte(`
        {
            "parallel": false,
            INVALID{
        }`)),
	}

	h := Batch(NewMux())

	response, err := h(context.Background(), event)

	if err == nil {
		t.Error("Expecting an error")
	}

	if response.FlowID != event.FlowID {
		t.Errorf("response.FlowID == %s, wants: %s", response.FlowID, event.FlowID)
	}

	if response.Name != "error" {
		t.Errorf(`response.Name == "%s", wants "error"`, response.Name)
	}
}

func Test_NewErrorWithMetadata(t *testing.T) {
	metadata := struct {
		SomeField  string `json:"field1"`
		OtherField int64  `json:"otherField"`
	}{
		SomeField:  "bla bla",
		OtherField: 42,
	}

	flowID := RandomID()

	event := NewErrorWithMetadata(flowID, "some error", metadata)

	if event.FlowID != flowID {
		t.Errorf(`event.FlowID = "%s", wants "%s"`, event.FlowID, flowID)
	}

	if event.Name != "error" {
		t.Errorf(`event.Name = "%s", wants: "%s"`, event.Name, "error")
	}

	readMetadata := map[string]interface{}{}

	err := json.Unmarshal(event.Metadata, &readMetadata)

	if err != nil {
		t.Errorf(`Error not expected: "%s"`, err.Error())
	}

	if readMetadata["field1"] != "bla bla" {
		t.Errorf(`readMetadata["field1"] = %s, wants "%s"`,
			readMetadata["field1"],
			"bla bla")
	}

	if readMetadata["otherField"] != 42.0 {
		t.Errorf(`readMetadata["otherField"] = %d, wants "%d"`,
			readMetadata["otherField"],
			42)
	}
}
