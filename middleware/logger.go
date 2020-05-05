package middleware

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/snicol/rf"

	"github.com/sirupsen/logrus"
	"github.com/snicol/yael"
)

const LoggerKey = "logger"

func Logger(logger *logrus.Logger) rf.MiddlewareFunc {
	return func(next rf.HandlerFunc) rf.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			entry := logger.WithFields(logrus.Fields{
				"http_method": r.Method,
				"http_path":   r.URL.Path,
			})

			start := time.Now()
			err := next(w, r)
			end := time.Now()

			dur := end.Sub(start)

			entry = entry.WithFields(logrus.Fields{
				"req_duration_us": dur.Microseconds(),
			})

			if err == nil {
				entry.WithFields(logrus.Fields{
					"http_status_code": http.StatusOK,
				}).Info("request handled")

				return nil
			}

			yaelErr, ok := err.(*yael.E)
			if !ok {
				entry.WithFields(logrus.Fields{
					"error": err.Error(),
				}).Error("internal server error")

				return err
			}

			ym, jErr := json.Marshal(yaelErr.Meta)
			if jErr != nil {
				return jErr
			}

			entry.WithFields(logrus.Fields{
				"code":             yaelErr.Code,
				"meta":             string(ym),
				"http_status_code": yael.StatusCode(*yaelErr),
			}).Warn(yaelErr.Code)

			return err
		}
	}
}
