package rf

import (
	"net/http"
)

// HandlerFunc defines the standard handler function type, used in the Handler
// interface. All custom handlers must return a func of this type to handle
// requests
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// ErrorHandlerFunc is the error handler function definition which is used in
// the Handler interface. All custom handlers must return a func of this type
// to handle errors
type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)

// MiddlewareFunc is the type used for all rf specific middleware
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// HandlerGroup holds middleware and is intended to group handlers of the same
// request type so that the same middleware stack can be applied
type HandlerGroup struct {
	mws []MiddlewareFunc
}

// NewHandlerGroup returns an instance of HandlerGroup with the middlware
// functions attached
func NewHandlerGroup(mws []MiddlewareFunc, more ...MiddlewareFunc) *HandlerGroup {
	return &HandlerGroup{
		mws: append(mws, more...),
	}
}

// Handler is the primary interface of this package - it defines the two
// standard functions to implement custom handler types
type Handler interface {
	Handle() HandlerFunc
	Error() ErrorHandlerFunc
}

// Use returns a http.Handler func, wrapped with middlewares which the request
// will be handled by the given rf.Handler
func (hs *HandlerGroup) Use(h Handler, mws ...MiddlewareFunc) http.HandlerFunc {
	fn := h.Handle()

	for _, mw := range mws {
		fn = mw(fn)
	}

	for _, mw := range hs.mws {
		fn = mw(fn)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err == nil {
			return
		}

		h.Error()(w, r, err)
		return
	}
}
