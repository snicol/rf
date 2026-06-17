// Package main demonstrates basic usage of the rf/basic handler type.
package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/snicol/rf"
	"github.com/snicol/rf/basic"
	"github.com/snicol/rf/middleware"
)

type Request struct {
	Input string `schema:"input,required"`
}

func Example(_ context.Context, req *Request) (*basic.Response, error) {
	return &basic.Response{
		Body:       "cheers for the input: " + req.Input,
		StatusCode: http.StatusOK,
		Headers: map[string]string{ // optional
			"Content-Type": "text/plain",
		},
	}, nil
}

func JSONEchoExample(_ context.Context, req *Request) (*basic.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err //nolint:wrapcheck // example code, wrapping not required
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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	g := rf.NewHandlerGroup(
		nil,
		middleware.Logger(logger),
		middleware.Recover(logger),
	)

	mux.Handle("/example", g.Use(basic.NewHandler(basic.GetParams, Example)))
	mux.Handle("/json_echo", g.Use(basic.NewHandler(basic.GetParams, JSONEchoExample)))
	mux.Handle("/post_form_json_echo", g.Use(basic.NewHandler(basic.PostForm, JSONEchoExample)))

	srv := &http.Server{
		Addr:         ":3003",
		Handler:      mux,
		ReadTimeout:  5 * time.Second, //nolint:mnd // example server timeouts
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second, //nolint:mnd // example server timeouts
	}
	log.Println(srv.ListenAndServe())
}
