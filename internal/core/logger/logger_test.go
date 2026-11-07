package core_logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestNewConfig(t *testing.T) {
	t.Setenv("LOGGER_LEVEL", "info")
	t.Setenv("LOGGER_FOLDER", t.TempDir())

	config, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	if config.Level != "info" || config.Folder == "" {
		t.Fatalf("NewConfig() = %+v", config)
	}
}

func TestNewLoggerWritesFileAndFiltersLevel(t *testing.T) {
	folder := t.TempDir()
	logger, err := NewLogger(Config{
		Level:  "info",
		Folder: folder,
	})
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	logger.Debug("hidden")
	logger.With(zap.String("component", "test")).Info("visible")
	if err := logger.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	files, err := filepath.Glob(filepath.Join(folder, "*.log"))
	if err != nil || len(files) != 1 {
		t.Fatalf("log files = %v, error = %v", files, err)
	}
	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(content), "visible") || !strings.Contains(string(content), "component") {
		t.Fatalf("log file does not contain message and field: %q", content)
	}
	if strings.Contains(string(content), "hidden") {
		t.Fatalf("log file contains debug message at info level: %q", content)
	}
}

func TestNewLoggerRejectsInvalidLevel(t *testing.T) {
	if _, err := NewLogger(Config{
		Level:  "not-a-level",
		Folder: t.TempDir(),
	}); err == nil {
		t.Fatal("NewLogger() accepted invalid level")
	}
}
