package core_pgx_pool

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

type Pool struct {
	*pgxpool.Pool
	opTimeout time.Duration
}

var _ core_postgres_pool.Pool = (*Pool)(nil)

func NewPool(ctx context.Context, config Config) (*Pool, error) {
	poolConfig, err := parsePoolConfig(config)
	if err != nil {
		return nil, err
	}

	connection, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create postgres connection pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()
	if err := connection.Ping(pingCtx); err != nil {
		connection.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Pool{
		Pool:      connection,
		opTimeout: config.Timeout,
	}, nil
}

func (p *Pool) Query(ctx context.Context, sql string, args ...any) (core_postgres_pool.Rows, error) {
	rows, err := p.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, mapError(err)
	}
	return pgxRows{Rows: rows}, nil
}

func (p *Pool) Begin(ctx context.Context) (core_postgres_pool.Tx, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return nil, mapError(err)
	}
	return pgxTx{Tx: tx}, nil
}

func (p *Pool) QueryRow(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
	return pgxRow{Row: p.Pool.QueryRow(ctx, sql, args...)}
}

func (p *Pool) Exec(ctx context.Context, sql string, args ...any) (core_postgres_pool.CommandTag, error) {
	tag, err := p.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, mapError(err)
	}
	return pgxCommandTag{CommandTag: tag}, nil
}

func (p *Pool) OpTimeout() time.Duration {
	return p.opTimeout
}

func parsePoolConfig(config Config) (*pgxpool.Config, error) {
	address := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(config.User, config.Password),
		Host:   net.JoinHostPort(config.Host, config.Port),
		Path:   "/" + config.Database,
	}
	query := address.Query()
	query.Set("sslmode", "disable")
	address.RawQuery = query.Encode()

	poolConfig, err := pgxpool.ParseConfig(address.String())
	if err != nil {
		return nil, fmt.Errorf("parse postgres connection config: %w", err)
	}
	return poolConfig, nil
}
