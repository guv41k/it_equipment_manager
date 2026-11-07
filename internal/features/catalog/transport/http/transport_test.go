package catalog_transport_http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	core_http_middleware "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/middleware"
	core_http_server "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/server"
)

type catalogServiceStub struct {
	called     string
	id         int64
	category   domain.Category
	location   domain.Location
	serviceErr error
}

func (s *catalogServiceStub) CreateCategory(_ context.Context, name string, description *string) (domain.Category, error) {
	s.called = "create category"
	category := domain.Category{
		Name:        name,
		Description: description,
	}
	s.category = category
	category.ID = 1
	return category, s.serviceErr
}

func (s *catalogServiceStub) GetCategories(context.Context) ([]domain.Category, error) {
	s.called = "list categories"
	if s.category.ID == 0 {
		return nil, s.serviceErr
	}
	return []domain.Category{s.category}, s.serviceErr
}

func (s *catalogServiceStub) GetCategory(_ context.Context, id int64) (domain.Category, error) {
	s.called = "get category"
	s.id = id
	return s.category, s.serviceErr
}

func (s *catalogServiceStub) UpdateCategory(_ context.Context, id int64, category domain.Category) (domain.Category, error) {
	s.called = "update category"
	s.id = id
	s.category = category
	category.ID = id
	return category, s.serviceErr
}

func (s *catalogServiceStub) CreateLocation(_ context.Context, name string, description *string) (domain.Location, error) {
	s.called = "create location"
	location := domain.Location{
		Name:        name,
		Description: description,
	}
	s.location = location
	location.ID = 2
	return location, s.serviceErr
}

func (s *catalogServiceStub) GetLocations(context.Context) ([]domain.Location, error) {
	s.called = "list locations"
	if s.location.ID == 0 {
		return nil, s.serviceErr
	}
	return []domain.Location{s.location}, s.serviceErr
}

func (s *catalogServiceStub) GetLocation(_ context.Context, id int64) (domain.Location, error) {
	s.called = "get location"
	s.id = id
	return s.location, s.serviceErr
}

func (s *catalogServiceStub) UpdateLocation(_ context.Context, id int64, location domain.Location) (domain.Location, error) {
	s.called = "update location"
	s.id = id
	s.location = location
	location.ID = id
	return location, s.serviceErr
}

func testCatalogRouter(t *testing.T, service *catalogServiceStub) http.Handler {
	t.Helper()
	log, err := core_logger.NewLogger(core_logger.Config{
		Level:  "error",
		Folder: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = log.Close() })
	router := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion1)
	router.AddRoutes(NewCategoriesHTTPHandler(service).Routes()...)
	router.AddRoutes(NewLocationsHTTPHandler(service).Routes()...)
	mux := http.NewServeMux()
	for pattern, handler := range router.Handlers() {
		mux.Handle(pattern, handler)
	}
	return core_http_middleware.Logger(log)(mux)
}

func sendCatalogRequest(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	handler.ServeHTTP(response, request)
	return response
}

func TestCatalogRoutes(t *testing.T) {
	tests := []struct {
		method string
		path   string
		body   string
		called string
		status int
		name   string
	}{
		{
			method: http.MethodPost,
			path:   "/api/v1/equipment-categories",
			body:   `{"name":"Ноутбук"}`,
			called: "create category",
			status: http.StatusCreated,
			name:   "Ноутбук",
		},
		{
			method: http.MethodGet,
			path:   "/api/v1/equipment-categories",
			called: "list categories",
			status: http.StatusOK,
		},
		{
			method: http.MethodGet,
			path:   "/api/v1/equipment-categories/1",
			called: "get category",
			status: http.StatusOK,
		},
		{
			method: http.MethodPut,
			path:   "/api/v1/equipment-categories/1",
			body:   `{"name":"Монитор"}`,
			called: "update category",
			status: http.StatusOK,
			name:   "Монитор",
		},
		{
			method: http.MethodPost,
			path:   "/api/v1/locations",
			body:   `{"name":"Склад"}`,
			called: "create location",
			status: http.StatusCreated,
			name:   "Склад",
		},
		{
			method: http.MethodGet,
			path:   "/api/v1/locations",
			called: "list locations",
			status: http.StatusOK,
		},
		{
			method: http.MethodGet,
			path:   "/api/v1/locations/2",
			called: "get location",
			status: http.StatusOK,
		},
		{
			method: http.MethodPut,
			path:   "/api/v1/locations/2",
			body:   `{"name":"Резервный склад"}`,
			called: "update location",
			status: http.StatusOK,
			name:   "Резервный склад",
		},
	}
	for _, test := range tests {
		t.Run(test.called, func(t *testing.T) {
			service := &catalogServiceStub{
				category: domain.Category{
					ID:   1,
					Name: "Ноутбук",
				},
				location: domain.Location{
					ID:   2,
					Name: "Склад",
				},
			}
			response := sendCatalogRequest(testCatalogRouter(t, service), test.method, test.path, test.body)
			if response.Code != test.status || service.called != test.called {
				t.Fatalf("status = %d, called = %q, body = %s", response.Code, service.called, response.Body.String())
			}
			if response.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("Content-Type = %q", response.Header().Get("Content-Type"))
			}
			if test.body != "" && !strings.Contains(response.Body.String(), `"name"`) {
				t.Fatalf("response = %s", response.Body.String())
			}
			if test.name != "" {
				gotName := service.category.Name
				if strings.Contains(test.path, "/locations") {
					gotName = service.location.Name
				}
				if gotName != test.name || !strings.Contains(response.Body.String(), test.name) {
					t.Fatalf("name passed to service = %q, response = %s", gotName, response.Body.String())
				}
			}
		})
	}
}

func TestCatalogEmptyList(t *testing.T) {
	service := &catalogServiceStub{}
	response := sendCatalogRequest(testCatalogRouter(t, service), http.MethodGet, "/api/v1/locations", "")
	if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
		t.Fatalf("empty list = %d %s", response.Code, response.Body.String())
	}
}

func TestCatalogRejectsAdditionalJSON(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		status int
		called string
	}{
		{
			name:   "second JSON object",
			method: http.MethodPost,
			path:   "/api/v1/equipment-categories",
			body:   `{"name":"Ноутбук"} {"name":"Монитор"}`,
			status: http.StatusBadRequest,
		},
		{
			name:   "trailing non-JSON text",
			method: http.MethodPut,
			path:   "/api/v1/locations/2",
			body:   `{"name":"Склад"} unexpected`,
			status: http.StatusBadRequest,
		},
		{
			name:   "trailing whitespace",
			method: http.MethodPost,
			path:   "/api/v1/equipment-categories",
			body:   "{\"name\":\"Ноутбук\"} \n\t",
			status: http.StatusCreated,
			called: "create category",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &catalogServiceStub{}
			response := sendCatalogRequest(testCatalogRouter(t, service), test.method, test.path, test.body)
			if response.Code != test.status || service.called != test.called {
				t.Fatalf("status = %d, service called = %q, response = %s", response.Code, service.called, response.Body.String())
			}
			if test.status == http.StatusBadRequest && !strings.Contains(response.Body.String(), `"error":"invalid_request"`) {
				t.Fatalf("invalid request response = %s", response.Body.String())
			}
		})
	}
}

func TestCatalogValidatesRequestAndMapsErrors(t *testing.T) {
	service := &catalogServiceStub{}
	router := testCatalogRouter(t, service)
	response := sendCatalogRequest(router, http.MethodPost, "/api/v1/equipment-categories", `{"name":"  "}`)
	if response.Code != http.StatusBadRequest || service.called != "" {
		t.Fatalf("invalid request: status = %d, called = %q", response.Code, service.called)
	}
	response = sendCatalogRequest(router, http.MethodGet, "/api/v1/locations/0", "")
	if response.Code != http.StatusBadRequest || service.called != "" {
		t.Fatalf("invalid ID: status = %d, called = %q", response.Code, service.called)
	}
	for _, test := range []struct {
		err    error
		status int
		code   string
	}{
		{
			err:    core_errors.ErrNotFound,
			status: http.StatusNotFound,
			code:   "not_found",
		},
		{
			err:    core_errors.ErrConflict,
			status: http.StatusConflict,
			code:   "conflict",
		},
		{
			err:    errors.New("private database detail"),
			status: http.StatusInternalServerError,
			code:   "internal_error",
		},
	} {
		service.serviceErr = test.err
		response = sendCatalogRequest(router, http.MethodGet, "/api/v1/locations/2", "")
		if response.Code != test.status || !strings.Contains(response.Body.String(), `"error":"`+test.code+`"`) {
			t.Fatalf("error response = %d %s", response.Code, response.Body.String())
		}
		if strings.Contains(response.Body.String(), "private database detail") {
			t.Fatal("internal error leaked to client")
		}
	}
}
