# rf

Request Framework - `rf` is a very light set of interfaces and handler types
that I've used on personal projects. After building a few things with the same
pieces of code strewn around I decided to standardise the request/response
patterns.

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
    g := rf.NewHandlerGroup(rpc.DefaultMiddleware(), middleware.Logger(logger))

    // RPC:
    // wrap the Example handler func (see below example) with a schema loaded using
    // gojsonschema.NewStringLoader/NewBytesLoader/etc, using the group we
    // just created
    http.Handle("/example", g.Use(rpc.NewHandler(Example, schema)))

    // pass in extra middleware per handler:
    http.Handle("/example", g.Use(rpc.NewHandler(Example, schema), MyCustomMiddleware, AnotherMiddleware, ...etc))

    // Basic: same applies
    http.Handle("/example_get", g.Use(basic.NewHandler(basic.GetParams, Example)), MyCustomMiddleware)
    http.Handle("/example_post", g.Use(basic.NewHandler(basic.PostForm, Example)), MyCustomMiddleware)

    // listen!
    http.ListenAndServe(":8000")
}
```

### RPC handlers

RPC handler function signatures are like so:

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

* Body
* StatusCode
* Headers

Technically all of these are optional, but once set will return the correct response to the client.

If no status code is set, it default to 200. If no headers are set, it default
to `Content-Type: text/plain` only.

## Notes

The next things that I will be adding are tests, a few more useful
middlewares (including moving shared logic from handlers into middleware). A lot
more of the code will be commented when it's more concrete.

**I will continue to chop and change features over time as my personal projects
evolve. There are no safety or security guarantees with this software!**
