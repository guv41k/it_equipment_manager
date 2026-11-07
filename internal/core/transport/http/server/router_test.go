package core_http_server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	core_http_middleware "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/middleware"
)

func TestAPIVersionRouter(t *testing.T) {
	server := newTestServer(t, &fakeReadinessChecker{
		ping: func(context.Context) error {
			return nil
		},
		timeout: time.Second,
	})
	router := NewAPIVersionRouter(ApiVersion1)
	router.AddRoutes(NewRoute(http.MethodGet, "/equipment/{id}", func(w http.ResponseWriter, r *http.Request) {
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

func TestRouterAndRouteMiddlewareOrder(t *testing.T) {
	server := newTestServer(t, &fakeReadinessChecker{
		ping: func(context.Context) error {
			return nil
		},
		timeout: time.Second,
	})
	var order []string
	mark := func(name string) core_http_middleware.Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}
	router := NewAPIVersionRouter(ApiVersion1, mark("version"))
	router.AddRoutes(Route{
		Method: http.MethodGet,
		Path:   "/equipment",
		Handler: func(w http.ResponseWriter, _ *http.Request) {
			order = append(order, "handler")
			w.WriteHeader(http.StatusNoContent)
		},
		Middleware: []core_http_middleware.Middleware{mark("route")},
	})
	server.RegisterAPIRouters(router)
	response := httptest.NewRecorder()
	server.mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/equipment", nil))
	if response.Code != http.StatusNoContent || len(order) != 3 || order[0] != "version" || order[1] != "route" || order[2] != "handler" {
		t.Fatalf("status = %d, middleware order = %v", response.Code, order)
	}
}
