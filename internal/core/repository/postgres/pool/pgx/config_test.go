package core_pgx_pool

import (
	"os"
	"strings"
	"testing"
	"time"
)

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "")
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("Unsetenv(%q): %v", key, err)
	}
}

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("POSTGRES_USER", "managerapp")
	t.Setenv("POSTGRES_PASSWORD", "test-password")
	t.Setenv("POSTGRES_DB", "managerapp")
}

func TestNewConfigDefaults(t *testing.T) {
	setRequiredEnv(t)
	unsetEnv(t, "POSTGRES_HOST")
	unsetEnv(t, "POSTGRES_PORT")
	unsetEnv(t, "POSTGRES_TIMEOUT")

	config, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	if config.Host != "127.0.0.1" || config.Port != "5432" || config.Timeout != 5*time.Second {
		t.Fatalf("unexpected connection settings: host=%q port=%q timeout=%s", config.Host, config.Port, config.Timeout)
	}
}

func TestNewConfigReadsOverrides(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("POSTGRES_HOST", "postgres.internal")
	t.Setenv("POSTGRES_PORT", "5433")
	t.Setenv("POSTGRES_TIMEOUT", "2s")

	config, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	if config.Host != "postgres.internal" || config.Port != "5433" || config.Timeout != 2*time.Second {
		t.Fatalf("unexpected connection settings: host=%q port=%q timeout=%s", config.Host, config.Port, config.Timeout)
	}
}

func TestNewConfigRejectsInvalidSettings(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "blank user",
			key:   "POSTGRES_USER",
			value: " ",
		},
		{
			name:  "blank password",
			key:   "POSTGRES_PASSWORD",
			value: " ",
		},
		{
			name:  "invalid port",
			key:   "POSTGRES_PORT",
			value: "65536",
		},
		{
			name:  "zero timeout",
			key:   "POSTGRES_TIMEOUT",
			value: "0s",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setRequiredEnv(t)
			t.Setenv("POSTGRES_HOST", "127.0.0.1")
			t.Setenv("POSTGRES_PORT", "5432")
			t.Setenv("POSTGRES_TIMEOUT", "5s")
			t.Setenv(test.key, test.value)

			if _, err := NewConfig(); err == nil {
				t.Fatalf("NewConfig() accepted %s=%q", test.key, test.value)
			}
		})
	}
}

func TestNewConfigDoesNotExposePasswordInError(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("POSTGRES_PORT", "invalid")

	_, err := NewConfig()
	if err == nil {
		t.Fatal("NewConfig() accepted an invalid port")
	}
	if strings.Contains(err.Error(), "test-password") {
		t.Fatalf("NewConfig() error exposes the password: %v", err)
	}
}
