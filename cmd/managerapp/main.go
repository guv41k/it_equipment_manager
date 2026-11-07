package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
	core_http_server "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/server"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context) (err error) {
	loggerConfig, err := core_logger.NewConfig()
	if err != nil {
		err = fmt.Errorf("load logger config: %w", err)
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	logger, err := core_logger.NewLogger(loggerConfig)
	if err != nil {
		err = fmt.Errorf("create logger: %w", err)
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	defer func() {
		if err != nil {
			logger.Error("application failed", zap.Error(err))
		}
		if closeErr := logger.Close(); closeErr != nil {
			fmt.Fprintln(os.Stderr, "close logger:", closeErr)
			err = errors.Join(err, fmt.Errorf("close logger: %w", closeErr))
		}
	}()

	logger.Debug("initializing postgres connection pool")
	postgresConfig, err := core_postgres_pool.NewConfig()
	if err != nil {
		return fmt.Errorf("load postgres config: %w", err)
	}
	pool, err := core_postgres_pool.NewConnectionPool(ctx, postgresConfig)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pool.Close()

	logger.Info("postgres connection ready", zap.String("database", postgresConfig.Database))
	httpConfig, err := core_http_server.NewConfig()
	if err != nil {
		return fmt.Errorf("load HTTP config: %w", err)
	}
	httpServer := core_http_server.NewHTTPServer(httpConfig, logger, pool)
	if err := httpServer.Run(ctx); err != nil {
		return fmt.Errorf("run HTTP server: %w", err)
	}
	return nil
}
