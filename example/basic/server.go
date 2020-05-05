package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/snicol/rf"
	"github.com/snicol/rf/basic"
	"github.com/snicol/rf/middleware"

	"github.com/sirupsen/logrus"
)

type Request struct {
	Input string `schema:"input,required"`
}

func Example(ctx context.Context, req *Request) (*basic.Response, error) {
	return &basic.Response{
		Body:       "cheers for the input: " + req.Input,
		StatusCode: 200, // optional
		Headers: map[string]string{ // optional
			"Content-Type": "text/plain",
		},
	}, nil
}

func JSONEchoExample(ctx context.Context, req *Request) (*basic.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	return &basic.Response{
		Body: string(body),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func main() {
	mux := http.NewServeMux()
	logger := logrus.New()

	g := rf.NewHandlerGroup(nil, middleware.Logger(logger))

	mux.Handle("/example", g.Use(basic.NewHandler(basic.GetParams, Example)))
	mux.Handle("/json_echo", g.Use(basic.NewHandler(basic.GetParams, JSONEchoExample)))
	mux.Handle("/post_form_json_echo", g.Use(basic.NewHandler(basic.PostForm, JSONEchoExample)))

	log.Println(http.ListenAndServe(":3003", mux))
}
