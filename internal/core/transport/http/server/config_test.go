package core_http_server

import (
	"os"
	"testing"
	"time"
)

func TestNewConfigDefaults(t *testing.T) {
	for _, key := range []string{"HTTP_ADDR", "HTTP_SHUTDOWN_TIMEOUT"} {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("Unsetenv(%q): %v", key, err)
		}
	}

	config, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	if config.Addr != "127.0.0.1:8080" || config.ShutdownTimeout != 10*time.Second {
		t.Fatalf("NewConfig() = %+v", config)
	}
}

func TestNewConfigReadsOverrides(t *testing.T) {
	t.Setenv("HTTP_ADDR", "localhost:9090")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "3s")

	config, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	if config.Addr != "localhost:9090" || config.ShutdownTimeout != 3*time.Second {
		t.Fatalf("NewConfig() = %+v", config)
	}
}

func TestNewConfigRejectsInvalidSettings(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "missing port",
			key:   "HTTP_ADDR",
			value: "127.0.0.1",
		},
		{
			name:  "invalid port",
			key:   "HTTP_ADDR",
			value: "127.0.0.1:65536",
		},
		{
			name:  "zero shutdown timeout",
			key:   "HTTP_SHUTDOWN_TIMEOUT",
			value: "0s",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("HTTP_ADDR", "127.0.0.1:8080")
			t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "10s")
			t.Setenv(test.key, test.value)

			if _, err := NewConfig(); err == nil {
				t.Fatalf("NewConfig() accepted %s=%q", test.key, test.value)
			}
		})
	}
}
