package core_http_middleware

import (
	"net/http"
	"time"

	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	"go.uber.org/zap"
)

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusRecorder) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusRecorder) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}

func (w *statusRecorder) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func Trace() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			recorder := &statusRecorder{ResponseWriter: w}
			defer func() {
				status := recorder.status
				if !recorder.wroteHeader {
					status = http.StatusOK
				}
				core_logger.FromContext(r.Context()).Info("HTTP request",
					zap.String("method", r.Method),
					zap.Int("status_code", status),
					zap.Duration("duration", time.Since(started)),
				)
			}()
			next.ServeHTTP(recorder, r)
		})
	}
}
