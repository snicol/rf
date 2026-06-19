package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.32.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/snicol/rf"
)

// Recover returns middleware that recovers from panics and logs them.
func Recover(opts ...Option) rf.MiddlewareFunc {
	o := applyOptions(opts)

	return func(next rf.HandlerFunc) rf.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) (retErr error) {
			start := time.Now()

			defer func() {
				p := recover()
				if p == nil {
					return
				}

				ctx := r.Context()

				span := trace.SpanFromContext(ctx)
				span.RecordError(fmt.Errorf("%v", p))
				span.SetStatus(codes.Error, "panic recovered")

				o.logger.ErrorContext(ctx, "panic recovered",
					slog.String(string(semconv.HTTPRequestMethodKey), r.Method),
					slog.String(string(semconv.URLPathKey), r.URL.Path),
					slog.Int64("req_duration_us", time.Since(start).Microseconds()),
					slog.Int(string(semconv.HTTPResponseStatusCodeKey), http.StatusInternalServerError),
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
