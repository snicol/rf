# rf — Request Framework

[![Go Reference](https://pkg.go.dev/badge/github.com/snicol/rf.svg)](https://pkg.go.dev/github.com/snicol/rf)
[![Test](https://github.com/snicol/rf/actions/workflows/test.yml/badge.svg)](https://github.com/snicol/rf/actions/workflows/test.yml)
[![Lint](https://github.com/snicol/rf/actions/workflows/lint.yml/badge.svg)](https://github.com/snicol/rf/actions/workflows/lint.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/snicol/rf)](https://goreportcard.com/report/github.com/snicol/rf)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

`rf` is a lightweight Go library of interfaces and handler types for building
HTTP services on top of `net/http`. It standardises the request/response
patterns I kept re-implementing across personal projects, while staying small
enough to drop into any router.

Everything works with Go's `net/http` directly, meaning you can bring your own
middleware, use Chi or http.Mux! You can rework middleware to gain advantage of
error handling (see middleware folder) within `rf`.

There are two main handlers for requests:

* RPC (POST, json + jsonschema,
[snicol/yael](https://github.com/snicol/yael) for errors)
* Basic (GET params and/or POST forms)

All requests types can happily live in harmony together under the same
mux/router. This allows for things such as OAuth2 services to be built with RPC
methods along with the required GET requests as per spec.

## Examples

See `example/{basic,rpc}` for the best examples of how each type works.

### Basic usage

```go
func main() {
    // creates a handler group which will apply a default stack of middleware
    // Tracer must come before Logger and Recover so the span is in context
    g := rf.NewHandlerGroup(rpc.DefaultMiddleware(), middleware.Tracer(), middleware.Logger(), middleware.Recover())

    // RPC:
    // wrap the Example handler func (see below example) with a schema loaded using
    // gojsonschema.NewStringLoader/NewBytesLoader/etc, using the group we
    // just created
    http.Handle("/example", g.Use(rpc.NewHandler(Example, schema)))

    // pass in extra middleware per handler:
    http.Handle("/example", g.Use(rpc.NewHandler(Example, schema), MyCustomMiddleware, AnotherMiddleware, ...etc))

    // Basic: same applies
    http.Handle("/example_get", g.Use(basic.NewHandler(basic.GetParams, Example), MyCustomMiddleware))
    http.Handle("/example_post", g.Use(basic.NewHandler(basic.PostForm, Example), MyCustomMiddleware))

    // listen!
    http.ListenAndServe(":8000")
}
```

### RPC handlers

Function signatures are like so:

```go
func Example(ctx context.Context, req *SomeInputType) (*YourOutput, error)
```

Input is validated using jsonschema.

### Basic handlers

Basic requests can handle either GET params or POST forms with the same
signature:

```go
func Handler(ctx context.Context, req *SomeInputType) (*basic.Response, error)
```

The request input from either params or form data is decoded into
`SomeInputType` using [gorilla/schema](https://github.com/gorilla/schema),
therefore all fields need to have the correct struct tags.

They differ slightly from RPC requests in that the response must be of type
`*basic.Response`. This type has the following fields:

* Body (`string`)
* StatusCode (`int`)
* Headers (`map[string]string`)

All of these are optional, but once set will return the correct response to the
client.

If no status code is set, it default to 200. If no headers are set, it default
to `Content-Type: text/plain` only.

## Middleware

The `middleware` package ships a small stack of reusable middleware. Each is an
`rf.MiddlewareFunc`, so they compose with `NewHandlerGroup` and `Use`:

* `middleware.Logger()` - request logging via `slog`
* `middleware.Tracer()` - OpenTelemetry span per request (must come before
  `Logger` and `Recover` so the span is in context)
* `middleware.Recover()` - recovers from panics in downstream handlers
* `middleware.RPCRequestOnly()` - rejects non-RPC requests
* `middleware.ChiURLParams()` - exposes Chi URL params to handlers

`Logger`, `Tracer`, and `Recover` accept options such as
`middleware.WithLogger(...)` and `middleware.WithTracer(...)`. RPC handlers also
get `rpc.DefaultMiddleware()` as a convenient base stack.

## Notes

This is a personal library: I will continue to chop and change features over
time as my projects evolve. **There are no safety or security guarantees with
this software!**

## License

Released under the [MIT License](LICENSE).
