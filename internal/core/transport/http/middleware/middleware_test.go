package core_http_middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
)

func TestRequestID(t *testing.T) {
	tests := []struct {
		name     string
		supplied string
		preserve bool
	}{
		{
			name: "generated when missing",
		},
		{
			name:     "preserves valid value",
			supplied: "client-123",
			preserve: true,
		},
		{
			name:     "replaces invalid value",
			supplied: "bad id",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var contextID string
			handler := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				contextID = RequestIDFromContext(r.Context())
				w.WriteHeader(http.StatusNoContent)
			}))
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			if test.supplied != "" {
				request.Header.Set("X-Request-ID", test.supplied)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			id := response.Header().Get("X-Request-ID")
			if !validRequestID(id) || id != contextID {
				t.Fatalf("response ID = %q, context ID = %q", id, contextID)
			}
			if test.preserve && id != test.supplied {
				t.Fatalf("request ID changed: %q", id)
			}
			if !test.preserve && id == test.supplied {
				t.Fatalf("invalid request ID retained: %q", id)
			}
		})
	}
}

func testLogger(t *testing.T) (*core_logger.Logger, string) {
	t.Helper()
	folder := t.TempDir()
	logger, err := core_logger.NewLogger(core_logger.Config{
		Level:  "info",
		Folder: folder,
	})
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	return logger, folder
}

func readLog(t *testing.T, logger *core_logger.Logger, folder string) string {
	t.Helper()
	if err := logger.Close(); err != nil {
		t.Fatalf("logger.Close() error = %v", err)
	}
	files, err := filepath.Glob(filepath.Join(folder, "*.log"))
	if err != nil || len(files) != 1 {
		t.Fatalf("log files = %v, error = %v", files, err)
	}
	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	return string(content)
}

func TestTraceRecordsImplicitStatusWithoutQuery(t *testing.T) {
	logger, folder := testLogger(t)
	handler := ChainMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("ok"))
		}),
		RequestID(),
		Logger(logger),
		Trace(),
	)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/items?token=private", nil))
	logText := readLog(t, logger, folder)

	if response.Code != http.StatusOK || !strings.Contains(logText, `"status_code": 200`) {
		t.Fatalf("response status = %d, log = %q", response.Code, logText)
	}
	if !strings.Contains(logText, `"path": "/items"`) || strings.Contains(logText, "token=private") {
		t.Fatalf("log path is incorrect or contains query parameters: %q", logText)
	}
	if !strings.Contains(logText, response.Header().Get("X-Request-ID")) {
		t.Fatalf("request ID missing from log: %q", logText)
	}
}

func TestPanicReturnsInternalErrorAndLogsPanic(t *testing.T) {
	logger, folder := testLogger(t)
	handler := ChainMiddleware(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic("private detail")
		}),
		RequestID(),
		Logger(logger),
		Trace(),
		Panic(),
	)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/items", nil))
	logText := readLog(t, logger, folder)

	if response.Code != http.StatusInternalServerError || response.Body.String() != "{\"error\":\"internal_error\"}\n" {
		t.Fatalf("panic response = %d %q", response.Code, response.Body.String())
	}
	if response.Header().Get("Content-Type") != "application/json" || response.Header().Get("X-Request-ID") == "" {
		t.Fatalf("panic response headers = %v", response.Header())
	}
	if strings.Contains(response.Body.String(), "private detail") || !strings.Contains(logText, "private detail") || !strings.Contains(logText, `"status_code": 500`) {
		t.Fatalf("panic detail visibility or status log is wrong: %q", logText)
	}
}
