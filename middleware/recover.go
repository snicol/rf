package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/snicol/rf"
)

func Recover(logger *slog.Logger) rf.MiddlewareFunc {
	if logger == nil {
		logger = slog.Default()
	}
	return func(next rf.HandlerFunc) rf.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) (retErr error) {
			start := time.Now()
			defer func() {
				p := recover()
				if p == nil {
					return
				}

				logger.ErrorContext(r.Context(), "panic recovered",
					slog.String("http_method", r.Method),
					slog.String("http_path", r.URL.Path),
					slog.Int64("req_duration_us", time.Since(start).Microseconds()),
					slog.Int("http_status_code", http.StatusInternalServerError),
					slog.String("panic", fmt.Sprint(p)),
					slog.String("stack_trace", string(debug.Stack())),
				)

				w.WriteHeader(http.StatusInternalServerError)
				retErr = nil
			}()
			return next(w, r)
		}
	}
}
