package core_http_middleware

import (
	"net/http"

	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	"go.uber.org/zap"
)

func Logger(log *core_logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestLog := log.With(
				zap.String("request_id", RequestIDFromContext(r.Context())),
				zap.String("path", r.URL.Path),
			)
			ctx := core_logger.ToContext(r.Context(), requestLog)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
