package rpc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/snicol/yael"
	"github.com/xeipuuv/gojsonschema"

	"github.com/snicol/rf"
)

// Handle returns the rf.HandlerFunc for this RPC handler.
func (h *Handler[Req, Res]) Handle() rf.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("reading body: %w", err)
		}

		_ = r.Body.Close() //nolint:errcheck // body close errors after ReadAll are not actionable

		if err := validateBody(body, h.schema); err != nil {
			return err
		}

		var req Req

		if len(body) > 0 {
			if err := json.Unmarshal(body, &req); err != nil {
				return fmt.Errorf("unmarshaling body: %w", err)
			}
		}

		res, err := h.fn(r.Context(), req)
		if err != nil {
			return err
		}

		if res == *new(Res) {
			result(w, "", http.StatusNoContent, "")

			return nil
		}

		resp, err := json.Marshal(res)
		if err != nil {
			return fmt.Errorf("marshaling response: %w", err)
		}

		result(w, string(resp), http.StatusOK, defaultContentType)

		return nil
	}
}

func validateBody(body []byte, schema gojsonschema.JSONLoader) error {
	if schema == nil {
		return nil
	}

	bodyLoader := gojsonschema.NewBytesLoader(body)

	schemaRes, err := gojsonschema.Validate(schema, bodyLoader)
	if err != nil {
		return fmt.Errorf("validating schema: %w", err)
	}

	if schemaRes.Valid() {
		return nil
	}

	yErr := yael.New(yael.BadRequest)

	yErr.Meta = map[string]any{
		"schema_error": make([]map[string]any, len(schemaRes.Errors())),
	}

	for i, reason := range schemaRes.Errors() {
		//nolint:forcetypeassert,errcheck,revive // key was set two lines above
		seMeta := yErr.Meta["schema_error"].([]map[string]any)

		seMeta[i] = map[string]any{
			"description": reason.Description(),
			"field":       reason.Field(),
			"type":        reason.Type(),
		}
	}

	return yErr
}

func result(w http.ResponseWriter, body string, statusCode int, contentType string) {
	if contentType == "" {
		contentType = "text/plain"
	}

	w.Header().Add("Content-Type", contentType)
	w.WriteHeader(statusCode)

	_, _ = w.Write([]byte(body)) //nolint:errcheck // response write errors are not actionable
}
