package middleware

import (
	"errors"
	"net/http"

	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.32.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/snicol/yael"

	"github.com/snicol/rf"
)

// Tracer returns middleware that starts an OTel server span for each request and
// sets HTTP semantic convention attributes. It must be used before Logger and Recover
// so that the span is present in the request context when those middlewares run.
func Tracer(opts ...Option) rf.MiddlewareFunc {
	o := applyOptions(opts)

	return func(next rf.HandlerFunc) rf.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			ctx, span := o.tracer.Start(r.Context(), r.Method+" "+r.URL.Path, trace.WithSpanKind(trace.SpanKindServer))
			defer span.End()

			sr := &statusRecorder{ResponseWriter: w}
			err := next(sr, r.WithContext(ctx))

			span.SetAttributes(
				semconv.HTTPRequestMethodKey.String(r.Method),
				semconv.URLPath(r.URL.Path),
				semconv.HTTPResponseStatusCode(sr.statusCode()),
			)

			if err == nil {
				return nil
			}

			yaelErr := &yael.E{}
			if errors.As(err, &yaelErr) {
				span.SetStatus(codes.Error, yaelErr.Code)

				return err
			}

			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			return err
		}
	}
}
