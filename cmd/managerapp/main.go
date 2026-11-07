package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	core_pgx_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/middleware"
	core_http_server "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/server"
	catalog_postgres_repository "gitlab.com/guv41k/it_equipment_manager/internal/features/catalog/repository/postgres"
	catalog_service "gitlab.com/guv41k/it_equipment_manager/internal/features/catalog/service"
	catalog_transport_http "gitlab.com/guv41k/it_equipment_manager/internal/features/catalog/transport/http"
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
	postgresConfig, err := core_pgx_pool.NewConfig()
	if err != nil {
		return fmt.Errorf("load postgres config: %w", err)
	}
	pool, err := core_pgx_pool.NewPool(ctx, postgresConfig)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pool.Close()

	logger.Info("postgres connection ready", zap.String("database", postgresConfig.Database))
	logger.Debug("initializing feature", zap.String("feature", "catalog"))
	categoriesRepository := catalog_postgres_repository.NewCategoriesRepository(pool)
	categoriesService := catalog_service.NewCategoriesService(categoriesRepository)
	categoriesTransportHTTP := catalog_transport_http.NewCategoriesHTTPHandler(categoriesService)
	locationsRepository := catalog_postgres_repository.NewLocationsRepository(pool)
	locationsService := catalog_service.NewLocationsService(locationsRepository)
	locationsTransportHTTP := catalog_transport_http.NewLocationsHTTPHandler(locationsService)

	httpConfig, err := core_http_server.NewConfig()
	if err != nil {
		return fmt.Errorf("load HTTP config: %w", err)
	}
	httpServer := core_http_server.NewHTTPServer(
		httpConfig,
		logger,
		pool,
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Trace(),
		core_http_middleware.Panic(),
	)
	apiVersionRouterV1 := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion1)
	apiVersionRouterV1.AddRoutes(categoriesTransportHTTP.Routes()...)
	apiVersionRouterV1.AddRoutes(locationsTransportHTTP.Routes()...)
	httpServer.RegisterAPIRouters(apiVersionRouterV1)

	if err := httpServer.Run(ctx); err != nil {
		return fmt.Errorf("run HTTP server: %w", err)
	}
	return nil
}
