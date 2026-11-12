package core_pgx_pool

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

type pgxRows struct {
	pgx.Rows
}

func (r pgxRows) Scan(dest ...any) error {
	return mapError(r.Rows.Scan(dest...))
}

func (r pgxRows) Err() error {
	return mapError(r.Rows.Err())
}

type pgxRow struct {
	pgx.Row
}

func (r pgxRow) Scan(dest ...any) error {
	return mapError(r.Row.Scan(dest...))
}

type pgxCommandTag struct {
	pgconn.CommandTag
}

type pgxTx struct {
	pgx.Tx
}

func (tx pgxTx) QueryRow(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
	return pgxRow{Row: tx.Tx.QueryRow(ctx, sql, args...)}
}

func (tx pgxTx) Exec(ctx context.Context, sql string, args ...any) (core_postgres_pool.CommandTag, error) {
	tag, err := tx.Tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, mapError(err)
	}
	return pgxCommandTag{CommandTag: tag}, nil
}

func (tx pgxTx) Commit(ctx context.Context) error {
	return mapError(tx.Tx.Commit(ctx))
}

func (tx pgxTx) Rollback(ctx context.Context) error {
	return mapError(tx.Tx.Rollback(ctx))
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("scan row: %w", core_postgres_pool.ErrNoRows)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			return fmt.Errorf("postgres foreign key violation: %w", core_postgres_pool.ErrViolatesForeignKey)
		case "23505":
			return fmt.Errorf("postgres unique violation: %w", core_postgres_pool.ErrUniqueViolation)
		}
	}
	return fmt.Errorf("postgres operation: %w: %w", core_postgres_pool.ErrUnknown, err)
}
