package core_pgx_pool

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

func TestMapError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "no rows",
			err:  pgx.ErrNoRows,
			want: core_postgres_pool.ErrNoRows,
		},
		{
			name: "foreign key violation",
			err: &pgconn.PgError{
				Code: "23503",
			},
			want: core_postgres_pool.ErrViolatesForeignKey,
		},
		{
			name: "unique violation",
			err: &pgconn.PgError{
				Code: "23505",
			},
			want: core_postgres_pool.ErrUniqueViolation,
		},
		{
			name: "unknown error",
			err:  errors.New("connection lost"),
			want: core_postgres_pool.ErrUnknown,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := mapError(test.err); !errors.Is(got, test.want) {
				t.Fatalf("mapError(%v) = %v, want %v", test.err, got, test.want)
			}
		})
	}
	if mapError(nil) != nil {
		t.Fatal("mapError(nil) must return nil")
	}
}
