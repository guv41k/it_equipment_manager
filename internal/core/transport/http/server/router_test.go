package core_http_server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAPIVersionRouter(t *testing.T) {
	server := newTestServer(t, &fakeReadinessChecker{
		ping: func(context.Context) error {
			return nil
		},
		timeout: time.Second,
	})
	router := NewAPIVersionRouter(ApiVersion1)
	router.RegisterRoutes(NewRoute(http.MethodGet, "/equipment/{id}", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.PathValue("id")))
	}))
	server.RegisterAPIRouters(router)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "versioned route and path parameter",
			method:     http.MethodGet,
			path:       "/api/v1/equipment/42",
			wantStatus: http.StatusOK,
			wantBody:   "42",
		},
		{
			name:       "unversioned path is absent",
			method:     http.MethodGet,
			path:       "/equipment/42",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "method is checked",
			method:     http.MethodPost,
			path:       "/api/v1/equipment/42",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, test.path, nil)
			server.mux.ServeHTTP(response, request)
			if response.Code != test.wantStatus || (test.wantBody != "" && response.Body.String() != test.wantBody) {
				t.Fatalf("response = %d %q, want %d %q", response.Code, response.Body.String(), test.wantStatus, test.wantBody)
			}
		})
	}
}
