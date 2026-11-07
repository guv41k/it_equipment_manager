package core_http_server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
)

type fakeReadinessChecker struct {
	ping    func(context.Context) error
	timeout time.Duration
	calls   int
}

func (f *fakeReadinessChecker) Ping(ctx context.Context) error {
	f.calls++
	return f.ping(ctx)
}

func (f *fakeReadinessChecker) OpTimeout() time.Duration {
	return f.timeout
}

func newTestServer(t *testing.T, db ReadinessChecker) *HTTPServer {
	t.Helper()
	logger, err := core_logger.NewLogger(core_logger.Config{
		Level:  "error",
		Folder: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	t.Cleanup(func() {
		if err := logger.Close(); err != nil {
			t.Errorf("logger.Close() error = %v", err)
		}
	})
	return NewHTTPServer(Config{
		Addr:            "127.0.0.1:0",
		ShutdownTimeout: time.Second,
	}, logger, db)
}

func TestHealthLiveDoesNotCheckDatabase(t *testing.T) {
	db := &fakeReadinessChecker{
		ping: func(context.Context) error {
			t.Fatal("liveness checked the database")
			return nil
		},
		timeout: time.Second,
	}
	server := newTestServer(t, db)
	response := httptest.NewRecorder()
	server.mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/live", nil))

	if response.Code != http.StatusOK || response.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("liveness response = %d %q", response.Code, response.Body.String())
	}
	if response.Header().Get("Content-Type") != "application/json" || db.calls != 0 {
		t.Fatalf("liveness headers = %v, database calls = %d", response.Header(), db.calls)
	}
}

func TestHealthReady(t *testing.T) {
	tests := []struct {
		name     string
		ping     func(context.Context) error
		wantCode int
		wantBody string
	}{
		{
			name: "database available",
			ping: func(context.Context) error {
				return nil
			},
			wantCode: http.StatusOK,
			wantBody: "{\"status\":\"ready\"}\n",
		},
		{
			name: "database unavailable",
			ping: func(context.Context) error {
				return errors.New("private database detail")
			},
			wantCode: http.StatusServiceUnavailable,
			wantBody: "{\"status\":\"unavailable\"}\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := &fakeReadinessChecker{
				ping:    test.ping,
				timeout: time.Second,
			}
			server := newTestServer(t, db)
			response := httptest.NewRecorder()
			server.mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

			if response.Code != test.wantCode || response.Body.String() != test.wantBody {
				t.Fatalf("readiness response = %d %q", response.Code, response.Body.String())
			}
			if response.Header().Get("Content-Type") != "application/json" || db.calls != 1 {
				t.Fatalf("readiness headers = %v, database calls = %d", response.Header(), db.calls)
			}
		})
	}
}

func TestHealthReadyTimesOut(t *testing.T) {
	var pingErr error
	db := &fakeReadinessChecker{
		ping: func(ctx context.Context) error {
			<-ctx.Done()
			pingErr = ctx.Err()
			return pingErr
		},
		timeout: 10 * time.Millisecond,
	}
	server := newTestServer(t, db)
	response := httptest.NewRecorder()
	server.mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	if response.Code != http.StatusServiceUnavailable || !errors.Is(pingErr, context.DeadlineExceeded) {
		t.Fatalf("readiness response = %d, ping error = %v", response.Code, pingErr)
	}
}
