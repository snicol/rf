package middleware

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/snicol/rf"
)

// ChiURLParams merges URL params (/books/{id}) into the *http.Request params
// map for later use.
// NOTE: this overrides any other GET paramaters with the same key
func ChiURLParams() func(next rf.HandlerFunc) rf.HandlerFunc {
	return func(next rf.HandlerFunc) rf.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			rctx := chi.RouteContext(r.Context())
			if rctx == nil {
				return errors.New("no chi Context found on request ctx - is Chi being used on this route?")
			}

			values := r.URL.Query()

			for i := 0; i < len(rctx.URLParams.Keys); i++ {
				values.Set(rctx.URLParams.Keys[i], rctx.URLParams.Values[i])
			}

			r.URL.RawQuery = values.Encode()

			return next(w, r)
		}
	}
}
