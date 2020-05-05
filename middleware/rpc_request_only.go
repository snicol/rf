package middleware

import (
	"net/http"
	"strings"

	"github.com/snicol/rf"

	"github.com/snicol/yael"
)

// RPCRequestOnly limits all requests to conform to our RPC request standard
func RPCRequestOnly() func(next rf.HandlerFunc) rf.HandlerFunc {
	return func(next rf.HandlerFunc) rf.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			if r.Method != http.MethodPost {
				return yael.New(yael.MethodNotAllowed)
			}

			if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
				return yael.New(yael.UnprocessableEntity)
			}

			return next(w, r)
		}
	}
}
