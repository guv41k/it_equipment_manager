package core_pgx_pool

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestParsePoolConfigEscapesCredentials(t *testing.T) {
	config, err := parsePoolConfig(Config{
		Host:     "127.0.0.1",
		Port:     "5432",
		User:     "user@company",
		Password: "pass:/?#@word",
		Database: "equipment test",
		Timeout:  time.Second,
	})
	if err != nil {
		t.Fatalf("parsePoolConfig() error = %v", err)
	}
	if config.ConnConfig.User != "user@company" || config.ConnConfig.Password != "pass:/?#@word" || config.ConnConfig.Database != "equipment test" {
		t.Fatal("parsePoolConfig() changed credentials or database name")
	}
}

func TestNewPoolChecksConnection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	pool, err := NewPool(ctx, Config{
		Host:     "127.0.0.1",
		Port:     "5432",
		User:     "managerapp",
		Password: "test-password",
		Database: "managerapp",
		Timeout:  time.Second,
	})
	if pool != nil {
		pool.Close()
		t.Fatal("NewPool() returned a pool for a canceled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("NewPool() error = %v, want context.Canceled", err)
	}
}
