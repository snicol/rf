package rpc

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/snicol/rf"

	"github.com/snicol/yael"
	"github.com/xeipuuv/gojsonschema"
)

func (rpc *Handler) Handle() rf.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}

		bodyLoader := gojsonschema.NewBytesLoader(body)
		res, err := gojsonschema.Validate(rpc.schema, bodyLoader)
		if err != nil {
			return err
		}

		if !res.Valid() {
			err := yael.New(yael.BadRequest)

			err.Meta = map[string]interface{}{
				"schema_error": make([]map[string]interface{}, len(res.Errors())),
			}

			for i, reason := range res.Errors() {
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

		fn := rpc.fn

		v := reflect.ValueOf(fn)
		t := v.Type()

		req := reflect.New(t.In(1).Elem())
		err = json.NewDecoder(r.Body).Decode(req.Interface())
		if err != nil {
			return err
		}

		inputs := []reflect.Value{
			reflect.ValueOf(ctx),
			req,
		}

		out := v.Call(inputs)

		if err := out[1]; !err.IsNil() {
			return err.Interface().(error)
		}

		resp, err := json.Marshal(out[0].Interface())
		if err != nil {
			return err
		}

		result(w, string(resp), http.StatusOK, &defaultContentType)
		return nil
	}
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
