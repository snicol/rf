package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	semconv "go.opentelemetry.io/otel/semconv/v1.32.0"

	"github.com/snicol/yael"

	"github.com/snicol/rf"
)

const tracerName = "github.com/snicol/rf/middleware"

// Logger returns middleware that logs each request with timing and status information.
// It should be used after Tracer so that log records are correlated to the active span
// via the request context.
func Logger(opts ...Option) rf.MiddlewareFunc {
	o := applyOptions(opts)

	return func(next rf.HandlerFunc) rf.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			ctx := r.Context()
			sr := &statusRecorder{ResponseWriter: w}
			start := time.Now()
			err := next(sr, r)

			status := sr.statusCode()
			duration := time.Since(start).Microseconds()

			base := o.logger.With(
				slog.String(string(semconv.HTTPRequestMethodKey), r.Method),
				slog.String(string(semconv.URLPathKey), r.URL.Path),
				slog.Int64("req_duration_us", duration),
			)

			if err == nil {
				base.InfoContext(ctx, "request handled", slog.Int(string(semconv.HTTPResponseStatusCodeKey), status))

				return nil
			}

			yaelErr := &yael.E{}
			if errors.As(err, &yaelErr) {
				base.WarnContext(ctx, yaelErr.Code,
					slog.String("code", yaelErr.Code),
					slog.Any("meta", yaelErr.Meta),
					slog.Int(string(semconv.HTTPResponseStatusCodeKey), yael.StatusCode(*yaelErr)),
				)

				return err
			}

			base.ErrorContext(ctx, "internal server error", slog.String("error", err.Error()))

			return err
		}
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func (sr *statusRecorder) statusCode() int {
	if sr.status == 0 {
		return http.StatusOK
	}

	return sr.status
}
