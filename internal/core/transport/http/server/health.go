package core_http_server

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type ReadinessChecker interface {
	Ping(ctx context.Context) error
	OpTimeout() time.Duration
}

func (s *HTTPServer) registerHealthRoutes(db ReadinessChecker) {
	s.mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, _ *http.Request) {
		writeHealth(w, http.StatusOK, `{"status":"ok"}`)
	})
	s.mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), db.OpTimeout())
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			s.log.Warn("postgres readiness check failed", zap.Error(err))
			writeHealth(w, http.StatusServiceUnavailable, `{"status":"unavailable"}`)
			return
		}
		writeHealth(w, http.StatusOK, `{"status":"ready"}`)
	})
}

func writeHealth(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body + "\n"))
}
