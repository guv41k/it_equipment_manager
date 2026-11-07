package core_http_server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	"go.uber.org/zap"
)

type HTTPServer struct {
	mux    *http.ServeMux
	config Config
	log    *core_logger.Logger
}

func NewHTTPServer(config Config, log *core_logger.Logger, db ReadinessChecker) *HTTPServer {
	server := &HTTPServer{
		mux:    http.NewServeMux(),
		config: config,
		log:    log,
	}
	server.registerHealthRoutes(db)
	return server
}

func (s *HTTPServer) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return fmt.Errorf("listen HTTP on %s: %w", s.config.Addr, err)
	}
	return s.serve(ctx, listener)
}

func (s *HTTPServer) serve(ctx context.Context, listener net.Listener) error {
	defer listener.Close()

	server := &http.Server{
		Handler:           s.mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	serveDone := make(chan error, 1)
	go func() {
		serveDone <- server.Serve(listener)
	}()
	s.log.Info("HTTP server listening", zap.String("addr", listener.Addr().String()))

	select {
	case err := <-serveDone:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	case <-ctx.Done():
		s.log.Info("shutting down HTTP server")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
		<-serveDone
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}
	if err := <-serveDone; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP: %w", err)
	}
	s.log.Info("HTTP server stopped")
	return nil
}
