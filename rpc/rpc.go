package rpc

import (
	"context"

	"github.com/xeipuuv/gojsonschema"
)

var defaultContentType = "application/json"

// Handler is an instance of a JSON based, POST only, jsonschema validated
// handler. This handler type is very opinionated and uses the yael package for
// handling errors.
type Handler[Req any, Res comparable] struct {
	fn     RPCHandlerFunc[Req, Res]
	schema gojsonschema.JSONLoader
}

type RPCHandlerFunc[Req any, Res comparable] func(context.Context, Req) (Res, error)

// NewHandler returns a handler instance with the provided handler function and
// jsonschema to validate the request with.
// The 'fn' argument provided must be a function with a signature like so:
//     func Example(ctx context.Context, req RequestType) (ResponseType, error)
// Any mismatch against the above format will result in a panic
func NewHandler[Req any, Res comparable](
	fn RPCHandlerFunc[Req, Res],
	schema gojsonschema.JSONLoader,
) *Handler[Req, Res] {
	if schema != nil {
		_, err := gojsonschema.NewSchemaLoader().Compile(schema)
		if err != nil {
			panic(err)
		}
	}

	h := &Handler[Req, Res]{
		fn:     fn,
		schema: schema,
	}

	return h
}
