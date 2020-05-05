package basic

import (
	"context"
	"errors"
	"net/http"
	"reflect"

	"github.com/snicol/rf"

	"github.com/gorilla/schema"
)

type RequestType int

const (
	GetParams RequestType = iota
	PostForm
)

type Handler struct {
	reqType RequestType
	fn      interface{}
}

type Response struct {
	Body       string
	StatusCode int
	Headers    map[string]string
}

func NewHandler(reqType RequestType, fn interface{}) *Handler {
	err := validateHandler(fn)
	if err != nil {
		panic(err)
	}

	return &Handler{
		reqType: reqType,
		fn:      fn,
	}
}

var decoder = schema.NewDecoder()

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

		if err := out[1]; !err.IsNil() {
			return err.Interface().(error)
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
		w.Write([]byte(res.Body))

		return nil
	}
}

func (h *Handler) Error() rf.ErrorHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func (h *Handler) decode(in interface{}, r *http.Request) error {
	switch h.reqType {
	case GetParams:
		if r.Method != http.MethodGet {
			return errors.New("unsupported method")
		}

		return decoder.Decode(in, r.URL.Query())
	case PostForm:
		if r.Method != http.MethodPost {
			return errors.New("unsupported method")
		}

		err := r.ParseForm()
		if err != nil {
			return err
		}

		return decoder.Decode(in, r.PostForm)
	default:
		return errors.New("unsupported request type")
	}
}

func validateHandler(fn interface{}) error {
	v := reflect.ValueOf(fn)
	t := v.Type()

	var errorType = reflect.TypeOf((*error)(nil)).Elem()
	var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	var responseType = reflect.TypeOf((*Response)(nil)).Elem()

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

	if t.In(1).Kind() != reflect.Ptr {
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
