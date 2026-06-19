package middleware

import (
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type options struct {
	logger *slog.Logger
	tracer trace.Tracer
}

// Option configures middleware behavior.
type Option func(*options)

// WithLogger sets the slog logger. Defaults to slog.Default().
func WithLogger(l *slog.Logger) Option {
	return func(o *options) { o.logger = l }
}

// WithTracer sets the OTel tracer. Defaults to otel.Tracer(tracerName).
func WithTracer(t trace.Tracer) Option {
	return func(o *options) { o.tracer = t }
}

func applyOptions(opts []Option) options {
	o := options{}

	for _, opt := range opts {
		opt(&o)
	}

	if o.logger == nil {
		o.logger = slog.Default()
	}

	if o.tracer == nil {
		o.tracer = otel.Tracer(tracerName)
	}

	return o
}
