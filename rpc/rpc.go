package rpc

import (
	"github.com/xeipuuv/gojsonschema"
)

var defaultContentType = "application/json"

// Handler is an instance of a JSON based, POST only, jsonschema validated
// handler. This handler type is very opinionated and uses the yael package for
// handling errors.
type Handler struct {
	fn     interface{}
	schema gojsonschema.JSONLoader
}

// NewHandler returns a handler instance with the provided handler function and
// jsonschema to validate the request with.
// The 'fn' argument provided must be a function with a signature like so:
//     func Example(ctx context.Context, req *RequestType) (ResponseType, error)
// Any mismatch against the above format will result in a panic
func NewHandler(fn interface{}, schema gojsonschema.JSONLoader) *Handler {
	err := validateHandler(fn)
	if err != nil {
		panic(err)
	}

	if err := validateSchema(schema); err != nil {
		panic(err)
	}

	return &Handler{
		fn:     fn,
		schema: schema,
	}
}
