package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/snicol/rf"
	"github.com/snicol/yael"
)

const LoggerKey = "logger"

func Logger(logger *slog.Logger) rf.MiddlewareFunc {
	if logger == nil {
		logger = slog.Default()
	}
	return func(next rf.HandlerFunc) rf.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			sr := &statusRecorder{ResponseWriter: w}

			start := time.Now()
			err := next(sr, r)

			base := logger.With(
				slog.String("http_method", r.Method),
				slog.String("http_path", r.URL.Path),
				slog.Int64("req_duration_us", time.Since(start).Microseconds()),
			)

			if err == nil {
				base.Info("request handled", slog.Int("http_status_code", sr.statusCode()))
				return nil
			}

			yaelErr, ok := err.(*yael.E)
			if !ok {
				base.Error("internal server error", slog.String("error", err.Error()))
				return err
			}

			base.Warn(yaelErr.Code,
				slog.String("code", yaelErr.Code),
				slog.Any("meta", yaelErr.Meta),
				slog.Int("http_status_code", yael.StatusCode(*yaelErr)),
			)

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
