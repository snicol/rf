package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/snicol/rf"
	"github.com/snicol/yael"
	"github.com/xeipuuv/gojsonschema"
)

var defaultContentType = "application/json"

// Handler is an instance of a JSON based, POST only, jsonschema validated
// handler. This handler type is very opinionated and uses the yael package for
// handling errors.
type Handler struct {
	fn rf.HandlerFunc
}

// NewHandler returns a handler instance with the provided handler function and
// jsonschema to validate the request with.
// The 'fn' argument provided must be a function with a signature like so:
//     func Example(ctx context.Context, req RequestType) (ResponseType, error)
// Any mismatch against the above format will result in a panic
func NewHandler[Request any, Response comparable](
	fn func(ctx context.Context, req Request) (Response, error),
	schema gojsonschema.JSONLoader,
) *Handler {
	if schema != nil {
		err := validateSchema(schema)
		if err != nil {
			panic(err)
		}
	}

	wrappedFn := func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}

		bodyLoader := gojsonschema.NewBytesLoader(body)
		schemaRes, err := gojsonschema.Validate(schema, bodyLoader)
		if err != nil {
			return err
		}

		if !schemaRes.Valid() {
			err := yael.New(yael.BadRequest)

			err.Meta = map[string]interface{}{
				"schema_error": make([]map[string]interface{}, len(schemaRes.Errors())),
			}

			for i, reason := range schemaRes.Errors() {
				seMeta := err.Meta["schema_error"].([]map[string]interface{})
				seMeta[i] = map[string]interface{}{
					"description": reason.Description(),
					"field":       reason.Field(),
					"type":        reason.Type(),
				}
			}

			return err
		}

		r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		var req Request
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			return err
		}

		res, err := fn(ctx, req)
		if err != nil {
			return err
		}

		if res == *new(Response) {
			result(w, "", http.StatusNoContent, nil)
			return nil
		}

		resp, err := json.Marshal(res)
		if err != nil {
			return err
		}

		result(w, string(resp), http.StatusOK, &defaultContentType)
		return nil
	}

	return &Handler{
		fn:     wrappedFn,
	}
}
