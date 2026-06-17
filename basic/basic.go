// Package basic provides a simple HTTP handler type for form-based request/response handling.
package basic

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/snicol/rf"

	"github.com/gorilla/schema"
)

// RequestType indicates how the handler should decode the incoming HTTP request.
type RequestType int

const (
	// GetParams decodes the handler input from URL query parameters.
	GetParams RequestType = iota
	// PostForm decodes the handler input from a POST form body.
	PostForm
)

// Handler wraps a typed function and handles HTTP request decoding and response writing.
type Handler struct {
	reqType RequestType
	fn      any
}

// Response is returned by handler functions to control the HTTP response.
type Response struct {
	Body       string
	StatusCode int
	Headers    map[string]string
}

// NewHandler returns a Handler for the given request type and handler function.
func NewHandler(reqType RequestType, fn any) *Handler {
	err := validateHandler(fn)
	if err != nil {
		panic(err)
	}

	return &Handler{
		reqType: reqType,
		fn:      fn,
	}
}

var decoder = schema.NewDecoder() //nolint:gochecknoglobals // package-level decoder is stateless and safe to share

// Handle returns the rf.HandlerFunc for this handler.
func (h *Handler) Handle() rf.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		fn := h.fn

		v := reflect.ValueOf(fn)
		t := v.Type()

		req := reflect.New(t.In(1).Elem())

		err := h.decode(req.Interface(), r)
		if err != nil {
			return err
		}

		inputs := []reflect.Value{
			reflect.ValueOf(r.Context()),
			req,
		}

		out := v.Call(inputs)

		if errVal := out[1]; !errVal.IsNil() {
			//nolint:errcheck,revive,forcetypeassert // validateHandler ensures return type implements error
			return errVal.Interface().(error)
		}

		res, ok := out[0].Interface().(*Response)
		if !ok {
			return errors.New("invalid response type found, expected get.Response{}")
		}

		w.Header().Add("Content-Type", "text/plain")

		for k, v := range res.Headers {
			w.Header().Set(k, v)
		}

		statusCode := 200
		if res.StatusCode != 0 {
			statusCode = res.StatusCode
		}

		w.WriteHeader(statusCode)

		_, _ = w.Write([]byte(res.Body)) //nolint:errcheck // response write errors are not actionable

		return nil
	}
}

// Error returns the rf.ErrorHandlerFunc for this handler.
func (*Handler) Error() rf.ErrorHandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request, err error) {
		w.WriteHeader(http.StatusInternalServerError)

		_, _ = w.Write([]byte(err.Error())) //nolint:errcheck // response write errors are not actionable
	}
}

func (h *Handler) decode(in any, r *http.Request) error {
	switch h.reqType {
	case GetParams:
		if r.Method != http.MethodGet {
			return errors.New("unsupported method")
		}

		if err := decoder.Decode(in, r.URL.Query()); err != nil {
			return fmt.Errorf("decoding query params: %w", err)
		}

		return nil
	case PostForm:
		if r.Method != http.MethodPost {
			return errors.New("unsupported method")
		}

		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("parsing form: %w", err)
		}

		if err := decoder.Decode(in, r.PostForm); err != nil {
			return fmt.Errorf("decoding form: %w", err)
		}

		return nil
	default:
		return errors.New("unsupported request type")
	}
}

func validateHandler(fn any) error {
	v := reflect.ValueOf(fn)
	t := v.Type()

	var (
		errorType    = reflect.TypeOf((*error)(nil)).Elem()
		contextType  = reflect.TypeOf((*context.Context)(nil)).Elem()
		responseType = reflect.TypeOf((*Response)(nil)).Elem()
	)

	if t.Kind() != reflect.Func {
		return errors.New("handler must be a function")
	}

	if t.NumIn() != 2 {
		return errors.New("handler needs two inputs")
	}

	if t.NumOut() != 2 {
		return errors.New("must be two return arguments")
	}

	if !t.In(0).Implements(contextType) {
		return errors.New("must take context as first argument")
	}

	if t.In(1).Kind() != reflect.Pointer {
		return errors.New("requset arg must be a ptr")
	}

	if t.Out(0).Elem() != responseType {
		return errors.New("must return an response")
	}

	if !t.Out(1).Implements(errorType) {
		return errors.New("must return an error")
	}

	return nil
}
