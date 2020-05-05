package rpc

import (
	"context"
	"errors"
	"reflect"

	"github.com/xeipuuv/gojsonschema"
)

func validateHandler(fn interface{}) error {
	v := reflect.ValueOf(fn)
	t := v.Type()

	var errorType = reflect.TypeOf((*error)(nil)).Elem()
	var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()

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

	if !t.Out(1).Implements(errorType) {
		return errors.New("must return an error")
	}

	return nil
}

func validateSchema(schema gojsonschema.JSONLoader) error {
	if schema == nil {
		return nil
	}

	_, err := gojsonschema.NewSchemaLoader().Compile(schema)
	return err
}
