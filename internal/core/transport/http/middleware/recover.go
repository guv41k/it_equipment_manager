package core_http_middleware

import (
	"errors"
	"net/http"

	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	"go.uber.org/zap"
)

func Recover(log *core_logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := &statusRecorder{ResponseWriter: w}
			defer func() {
				if recovered := recover(); recovered != nil {
					if recoveredErr, ok := recovered.(error); ok && errors.Is(recoveredErr, http.ErrAbortHandler) {
						panic(recovered)
					}
					log.Error("HTTP handler panic",
						zap.String("request_id", RequestIDFromContext(r.Context())),
						zap.Any("panic", recovered),
						zap.Stack("stack"),
					)
					if recorder.wroteHeader {
						return
					}
					recorder.Header().Set("Content-Type", "application/json")
					recorder.Header().Set("Cache-Control", "no-store")
					recorder.WriteHeader(http.StatusInternalServerError)
					_, _ = recorder.Write([]byte("{\"error\":\"internal_error\"}\n"))
				}
			}()
			next.ServeHTTP(recorder, r)
		})
	}
}
