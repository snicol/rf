// Package main demonstrates usage of the rf/rpc handler type with jsonschema validation.
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/snicol/rf"
	"github.com/snicol/rf/middleware"
	"github.com/snicol/rf/rpc"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/xeipuuv/gojsonschema"
)

type Request struct {
	Input string `json:"input"`
}

type Response struct {
	Output string `json:"output"`
}

func Example(_ context.Context, req *Request) (*Response, error) {
	return &Response{
		Output: req.Input + " and some output.",
	}, nil
}

//nolint:gochecknoglobals // example global schema loader
var schema = gojsonschema.NewStringLoader(`{
	"type": "object",
	"additionalProperties": false,

	"required": ["input"],

	"properties": {
		"input": {
			"type": "string",
			"minLength": 1
		}
	}
}`)

func exampleMux() { //nolint:unused // demonstrates stdlib mux usage alongside the chi example in main
	mux := http.NewServeMux()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	g := rf.NewHandlerGroup(
		rpc.DefaultMiddleware(),
		middleware.Tracer(),
		middleware.Logger(),
		middleware.Recover(),
	)

	mux.Handle("/example", g.Use(rpc.NewHandler(Example, schema)))

	srv := &http.Server{
		Addr:         ":3003",
		Handler:      mux,
		ReadTimeout:  5 * time.Second, //nolint:mnd // example server timeouts
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second, //nolint:mnd // example server timeouts
	}
	log.Println(srv.ListenAndServe())
}

func main() {
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.StripSlashes)

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	g := rf.NewHandlerGroup(
		rpc.DefaultMiddleware(),
		middleware.Tracer(),
		middleware.Logger(),
		middleware.Recover(),
	)

	r.Post("/example", g.Use(rpc.NewHandler(Example, schema)))

	srv := &http.Server{
		Addr:         ":3003",
		Handler:      r,
		ReadTimeout:  5 * time.Second, //nolint:mnd // example server timeouts
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second, //nolint:mnd // example server timeouts
	}
	log.Println(srv.ListenAndServe())
}
