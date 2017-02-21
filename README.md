Events
======
An events protocol designed to be a standard communication between applications.
It is meant to be used synchronously (JSON over HTTP, REST) or asynchronously (WebSocket, Socket.IO)

This is a production-ready library, however we still working on the documentation and it is coupled with [Newrelic agent](https://github.com/newrelic/go-agent) and [xlog](https://github.com/rs/xlog)


Example:
--------
```javascript
{
	"name": "some event",
	"version": 42,
	"id": "2465a86f-3857-423e-af86-41f67880172f",
	"flowId": "d00a5c99-ea0e-4b39-bfdc-bf1028a9c95f",
	"payload": {
		"data1": "fsadfsdf",
		"an_int64": 65484548474984654,
		"some_float": 1864.4568
	},
	"metadata": {
		"origin": "Documentation",
		"originId": "RFC-GB 0001",
		"timestamp": 1482162952
	}
}
```

#### Schema
```javascript
{
	"$schema": "http://json-schema.org/draft-04/schema#",
	"type": "object",
	"properties": {
		"name": {
			"type": "string",
			"description": "Event name"
		},

		"version": {
			"type": "integer",
			"minimum": 1,
			"description": "Event version"
		},


	   "id": {
		   "type": "string",
		   "description": "Unique event identifier, suggestion: UUIDv4"
	   },

	   "flowId": {
		   "type": "string",
		   "description": "Unique Flow ID. It is intended to track the flow of information in a micro-service architecture"
	   },

		"payload": {
			"type": "object",
			"description": "All data this event needs"
		},

		"metadata": {
			"type": "object",
			"description": "Metadata about this event. It is not mandatory, even with an empty metadata the application/event must work"
		}
	},

	"required": [
		"name",
		"version",
		"id",
		"flowId",
		"payload",
	]
}

```

## Installing GO

Follow this guide https://golang.org/doc/install

Please install Go version >= 1.7

## Installing glide (dependency management)

Download the binary and put in the $PATH

https://github.com/Masterminds/glide/releases

## Install dependencies

```
$ glide install
```

## Testing

```
$ go test . `glide nv`
```

### Using Goconvey (Live Test Runner)

```
$ go get github.com/smartystreets/goconvey
```

```
$ $GOPATH/bin/goconvey
```

Access http://localhost:8080