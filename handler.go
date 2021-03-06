package events

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"sort"

	"github.com/alecthomas/jsonschema"
)

// Handler Interface for event handler
type Handler interface {
	Serve(context.Context, Event) (Event, error)
}

// EventDoc Interface that defines methods used to generete
// events documentation
type EventDoc interface {
	Example() (interface{}, interface{})
	Input() interface{}
	Output() interface{}
	Doc() string
}

// HandlerFunc a event handler
type HandlerFunc func(context.Context, Event) (Event, error)

// Serve implements Handler interface
func (h HandlerFunc) Serve(ctx context.Context, event Event) (Event, error) {
	return h(ctx, event)
}

// Mux Events mux
type Mux struct {
	events map[eventKey]Handler
	tracer HTTPTracker
}

type eventKey struct {
	Name    string
	Version int
}

// NewMux returns a new events Mux
func NewMuxWithTracker(tracer HTTPTracker) *Mux {
	return &Mux{
		events: map[eventKey]Handler{},
		tracer: tracer,
	}
}

// NewMuxNoOpTracker returns a new events Mux
func NewMux() *Mux {
	return &Mux{
		events: map[eventKey]Handler{},
		tracer: NewNoOpTracker(),
	}
}

// Add adds a HandlerFunc into the Mux
func (m *Mux) Add(name string, version int, handler Handler) {
	key := eventKey{name, version}
	m.events[key] = handler
}

func (m *Mux) get(name string, version int) (Handler, bool) {
	key := eventKey{name, version}
	h, ok := m.events[key]
	return h, ok
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	event := Event{}

	err := json.NewDecoder(r.Body).Decode(&event)

	if err != nil {
		ctx = m.tracer.NoticeError(ctx, err)

		json.NewEncoder(w).Encode(
			NewError(
				"",
				err.Error(),
			))
		return
	}

	handler, ok := m.get(event.Name, event.Version)

	if !ok {
		ctx = m.tracer.NoticeEventError(ctx, event, err)

		json.NewEncoder(w).Encode(
			NewError(
				event.FlowID,
				"Event not Found",
			))
		return
	}

	ctx = m.tracer.Start(ctx, event, w, r)
	response, err := handler.Serve(ctx, event)
	ctx = m.tracer.End(ctx, event, err)

	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		m.tracer.NoticeEventError(ctx, event, err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
	}
}

type doc struct {
	Name         string
	Version      int
	DocString    template.HTML
	InputSchema  string
	OutputSchema string

	InputExample  string
	OutputExample string

	HaveExtendedDoc bool
}

// ServeDoc - Serves all documentation
func (m *Mux) ServeDoc(w http.ResponseWriter, r *http.Request) {
	tpl := template.Must(template.New("doc").Parse(htmlTemplate))

	docs := []doc{}

	for key, handler := range m.events {
		eventDoc := doc{
			Name:    key.Name,
			Version: key.Version,
		}

		if eventWithDoc, ok := handler.(EventDoc); ok {
			eventDoc.HaveExtendedDoc = true

			input, _ := json.MarshalIndent(jsonschema.Reflect(eventWithDoc.Input()), "", "  ")
			eventDoc.InputSchema = string(input)

			output, _ := json.MarshalIndent(jsonschema.Reflect(eventWithDoc.Output()), "", "  ")
			eventDoc.OutputSchema = string(output)

			inputExample, outputExample := eventWithDoc.Example()

			jsonInputExample, _ := json.MarshalIndent(inputExample, "", "  ")
			eventDoc.InputExample = string(jsonInputExample)

			jsonOutputExample, _ := json.MarshalIndent(outputExample, "", "  ")
			eventDoc.OutputExample = string(jsonOutputExample)

			eventDoc.DocString = template.HTML(eventWithDoc.Doc())
		}
		docs = append(docs, eventDoc)
	}

	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Name < docs[j].Name
	})

	tpl.Execute(w, docs)
}

var htmlTemplate = `
<html>
<head>
    <title>Events documentation</title>
    <style>
        body {
            font-family: 'Open Sans', Arial, sans-serif;
            padding: 1em;
        }

        .title {
            font-family: 'Oswald', 'Impact', 'Arial Black', sans-serif;
            text-align: center;
        }

        .event {
            padding: 1em;
            border: 1px solid #AAA;
            margin-bottom: 1em;
        }

        .event__title {
            font-family: 'Inconsolata', 'Droid Sans Mono', 'Courier New', monospace;
            padding: 1em;
            font-size: large;
            background-color: #595b5d;
            margin-bottom: 1em;
            color: white;
        }

        pre {
            background-color: #CCC;
            color: #1d1f21;
            border: 1px solid #AAA;
            padding: 1em;
        }
        code {
        color: #444;
            background-color: #f5f5f5;
            font: monospace;
            padding: 1px 4px;
            border: 1px solid #cfcfcf;
            border-radius: 3px;
        }
    </style>
</head>
<body>
<h1 class="title">Events</h1>
<div>
    <ul>
        {{ range . }}        
            {{ if .HaveExtendedDoc }}
                <li><a href="#{{ .Name }}">{{ .Name }}</a></li>
            {{ else }}
                <li>{{ .Name }}</li>
            {{ end }}
        {{ end }}
    </ul>
    {{ range . }}
    <div class="event" id="{{ .Name }}">
        <div class="event__title">{{ .Name }} (Version: {{ .Version }})</div>

        {{ if .HaveExtendedDoc }}
        <p>{{ .DocString }}</p>

        <h2>Input Example</h2>

        <pre>{{ .InputExample }}</pre>

        <h2>Input Schema</h2>

        <pre>{{ .InputSchema }}</pre>

        <h2>Output Example</h2>

        <pre>{{ .OutputExample }}</pre>

        <h2>Output Schema</h2>

        <pre>{{ .OutputSchema }}</pre>
        {{ end }}
    </div>
    {{ end }}
</div>
</body>
</html>`
