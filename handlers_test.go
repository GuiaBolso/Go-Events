package events

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockHandlerFunc(_ context.Context, event Event) (Event, error) {
	return Event{
		FlowID:  event.FlowID,
		Name:    event.Name + ":response",
		ID:      RandomID(),
		Version: 1,
		Payload: event.Payload,
	}, nil
}

func mockEventError(_ context.Context, event Event) (Event, error) {
	return Event{
		FlowID:  event.FlowID,
		Name:    event.Name + ":error",
		ID:      RandomID(),
		Version: 1,
		Payload: event.Payload,
	}, errors.New("some error")
}

func mockHandlerFuncInvalidEvent(_ context.Context, event Event) (Event, error) {
	return Event{
		FlowID:  event.FlowID,
		Name:    event.Name + ":response",
		ID:      RandomID(),
		Version: 1,
		Payload: []byte("INVALID ["),
	}, nil
}

func Test_NewMux(t *testing.T) {
	mux := NewMux()

	if mux.events == nil {
		t.Error("Mux does not have an events map")
	}
}

func Test_Add(t *testing.T) {
	mux := NewMux()

	mux.Add("mock", 1, HandlerFunc(mockHandlerFunc))

	if len(mux.events) != 1 {
		t.Error("Events were not added corectly")
	}
}

func Test_get(t *testing.T) {
	mux := NewMux()
	key := eventKey{"mock", 1}

	mux.Add(key.Name, key.Version, HandlerFunc(mockHandlerFunc))

	_, ok := mux.get(key.Name, key.Version)

	if !ok {
		t.Error("Could not get event handler")
	}
}

func Test_ServerHTTP_invalid_Request(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/events/", strings.NewReader("INVALID [}"))

	mockTracker := &MockTracker{
		NoticeErrorFn: func(ctx context.Context, err error) context.Context { return ctx },
	}

	mux := NewMuxWithTracker(mockTracker)
	mux.ServeHTTP(w, r)

	body := w.Body.String()

	if w.Code != http.StatusOK {
		t.Errorf("w.Code = %d, expecting %d", w.Code, http.StatusOK)
	}

	if len(body) == 0 {
		t.Error("Body cannot be empty")
	}

	if mockTracker.NoticeErrorCount != 1 {
		t.Error("NoticeError must be called once")
	}
}

func Test_ServerHTTP_Event_not_found(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/events/", strings.NewReader(mockEvent))

	mockTracker := &MockTracker{
		NoticeEventErrorFn: func(ctx context.Context, event Event, err error) context.Context { return ctx },
	}

	mux := NewMuxWithTracker(mockTracker)

	mux.Add("some event", 1, HandlerFunc(mockHandlerFunc))
	mux.ServeHTTP(w, r)

	body := w.Body.String()

	if w.Code != http.StatusOK {
		t.Errorf("w.Code = %d, expecting %d", w.Code, http.StatusOK)
	}

	if len(body) == 0 {
		t.Error("Body cannot be empty")
	}

	if mockTracker.NoticeEventErrorCount != 1 {
		t.Error("NoticeEventError must be called once")
	}
}

func Test_ServerHTTP_event_handler_error(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/events/", strings.NewReader(mockEvent))

	mux := NewMux()

	mux.Add("some event", 42, HandlerFunc(mockEventError))
	mux.ServeHTTP(w, r)

	body := w.Body.String()

	if w.Code != http.StatusOK {
		t.Errorf("w.Code = %d, expecting %d", w.Code, http.StatusOK)
	}

	response := Event{}
	json.NewDecoder(w.Body).Decode(&response)

	if response.FlowID != "d00a5c99-ea0e-4b39-bfdc-bf1028a9c95f" {
		t.Error("Expecting the same flowID as the request event")
	}

	expectedName := "some event:error"
	if response.Name != expectedName {
		t.Errorf(`Got response.Name = %s, expecting: %s\n\nResponse: %s`,
			response.Name,
			expectedName,
			body)
	}
}

func Test_ServerHTTP(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/events/", strings.NewReader(mockEvent))

	mux := NewMux()

	mux.Add("some event", 42, HandlerFunc(mockHandlerFunc))
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("w.Code = %d, expecting %d", w.Code, http.StatusOK)
	}

	response := Event{}
	json.NewDecoder(w.Body).Decode(&response)

	if response.FlowID != "d00a5c99-ea0e-4b39-bfdc-bf1028a9c95f" {
		t.Error("Expecting the same flowID as the request event")
	}

	if response.Name != "some event:response" {
		t.Error("Expecting a different response")
	}
}

func Test_ServerHTTP_enconding_response_error(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/events/", strings.NewReader(mockEvent))

	mockTracker := &MockTracker{
		StartFn:            func(ctx context.Context, _ Event, _ http.ResponseWriter, _ *http.Request) context.Context { return ctx },
		NoticeEventErrorFn: func(ctx context.Context, _ Event, _ error) context.Context { return ctx },
		EndFn:              func(ctx context.Context, _ Event, _ error) context.Context { return ctx },
	}
	mux := NewMuxWithTracker(mockTracker)

	mux.Add("some event", 42, HandlerFunc(mockHandlerFuncInvalidEvent))
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("w.Code = %d, expecting %d", w.Code, http.StatusInternalServerError)
	}

	if mockTracker.NoticeEventErrorCount != 1 {
		t.Error("NoticeEventError must be called only once")
	}

	if mockTracker.EndCount != 1 {
		t.Error("End must be called only once")
	}

	if mockTracker.StartCount != 1 {
		t.Error("StartCount must be called only once")
	}
}

func Test_ServeDoc(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/doc", nil)

	ev := &mockEventStruct{}

	mux := NewMux()

	mux.Add("TestEvent", 42, ev)
	mux.Add("ZTestEvent", 1, ev)

	mux.ServeDoc(w, r)

	if w.Code != http.StatusOK {
		b, _ := ioutil.ReadAll(w.Body)
		t.Errorf(`w.Code == %d, wants: %d\nBody: "%s"`, w.Code, http.StatusOK, string(b))
	}

	b, _ := ioutil.ReadAll(w.Body)
	bodyStr := string(b)

	if !strings.Contains(bodyStr, "TestEvent") {
		t.Error("Could not find expected information on documentation")
	}
	if !strings.Contains(bodyStr, "f1") {
		t.Error("Could not find expected information on documentation")
	}

	if !strings.Contains(bodyStr, "mockEventStructInput") {
		t.Error("Could not find expected information on documentation")
	}

	if !strings.Contains(bodyStr, "f4") {
		t.Error("Could not find expected information on documentation")
	}
}

type mockEventStruct struct{}

type mockEventStructInput struct {
	Field1 string `json:"f1"`
	Field2 int64  `json:"someNumber"`
	Omit   string `json:"omit,omitempty"`
}

type mockEventStructOutput struct {
	Field3 string  `json:"field3"`
	Field4 float64 `json:"f4"`
}

func (h *mockEventStruct) Serve(context.Context, Event) (Event, error) {
	return Event{}, nil
}

func (h *mockEventStruct) Example() (interface{}, interface{}) {
	in := h.Input().(mockEventStructInput)
	in.Field1 = "some string"
	in.Field2 = 42

	out := h.Output().(mockEventStructOutput)

	out.Field3 = "another string"
	out.Field4 = 10.42

	return in, out
}

func (h *mockEventStruct) Input() interface{} {
	return mockEventStructInput{}
}

func (h *mockEventStruct) Output() interface{} {
	return mockEventStructOutput{}
}

func (h *mockEventStruct) Doc() string {
	return "Mock documentation"
}

var mockEvent = `
{
    "name": "some event",
    "version": 42,
    "id": "2465a86f-3857-423e-af86-41f67880172f",
    "flowId": "d00a5c99-ea0e-4b39-bfdc-bf1028a9c95f",
    "payload":{ 
        "data1": "fsadfsdf",
        "an_int64": 65484548474984654,
        "some_float": 1864.4568
    },
    "metadata": {
        "origin": "Documentation",
        "originId": "RFC-GB 0001",
        "timestamp": 1482162952
    }
}`
