package rpc

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/snicol/rf"
	"github.com/snicol/yael"
	"github.com/xeipuuv/gojsonschema"
)

func (h *Handler[Req, Res]) Handle() rf.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}

		if err := validateBody(body, h.schema); err != nil {
			return err
		}

		r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		var req Req
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			return err
		}

		res, err := h.fn(r.Context(), req)
		if err != nil {
			return err
		}

		if res == *new(Res) {
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
}

func validateBody(body []byte, schema gojsonschema.JSONLoader) error {
	if schema == nil {
		return nil
	}

	bodyLoader := gojsonschema.NewBytesLoader(body)
	schemaRes, err := gojsonschema.Validate(schema, bodyLoader)
	if err != nil {
		return err
	}

	if schemaRes.Valid() {
		return nil
	}

	yErr := yael.New(yael.BadRequest)

	yErr.Meta = map[string]interface{}{
		"schema_error": make([]map[string]interface{}, len(schemaRes.Errors())),
	}

	for i, reason := range schemaRes.Errors() {
		seMeta := yErr.Meta["schema_error"].([]map[string]interface{})
		seMeta[i] = map[string]interface{}{
			"description": reason.Description(),
			"field":       reason.Field(),
			"type":        reason.Type(),
		}
	}

	return yErr
}

func result(w http.ResponseWriter, body string, statusCode int, contentType *string) {
	var headerContentType = "text/plain"

	if contentType != nil {
		headerContentType = *contentType
	}

	w.Header().Add("Content-Type", headerContentType)
	w.WriteHeader(statusCode)
	w.Write([]byte(body))
}
