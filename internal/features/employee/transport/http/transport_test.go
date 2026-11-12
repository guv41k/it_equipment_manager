package employee_transport_http

import (
	"context"
	"encoding/json"
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

type employeesServiceStub struct {
	called   string
	id       int64
	employee domain.Employee
	err      error
}

func (s *employeesServiceStub) CreateEmployee(_ context.Context, fullName, department string) (domain.Employee, error) {
	s.called = "create"
	s.employee = domain.Employee{
		ID:         1,
		FullName:   fullName,
		Department: department,
		IsActive:   true,
	}
	return s.employee, s.err
}

func (s *employeesServiceStub) GetEmployees(context.Context) ([]domain.Employee, error) {
	s.called = "list"
	if s.employee.ID == 0 {
		return nil, s.err
	}
	return []domain.Employee{s.employee}, s.err
}

func (s *employeesServiceStub) GetEmployee(_ context.Context, id int64) (domain.Employee, error) {
	s.called = "get"
	s.id = id
	return s.employee, s.err
}

func (s *employeesServiceStub) UpdateEmployee(_ context.Context, id int64, employee domain.Employee) (domain.Employee, error) {
	s.called = "update"
	s.id = id
	s.employee = employee
	s.employee.ID = id
	return s.employee, s.err
}

func testEmployeesRouter(t *testing.T, service *employeesServiceStub) http.Handler {
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
	router.AddRoutes(NewEmployeesHTTPHandler(service).Routes()...)
	mux := http.NewServeMux()
	for pattern, handler := range router.Handlers() {
		mux.Handle(pattern, handler)
	}
	return core_http_middleware.Logger(log)(mux)
}

func sendEmployeesRequest(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	handler.ServeHTTP(response, request)
	return response
}

func TestEmployeeRoutes(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		called string
		status int
	}{
		{
			name:   "create",
			method: http.MethodPost,
			path:   "/api/v1/employees",
			body:   `{"full_name":"Иван Иванов","department":"ИТ"}`,
			called: "create",
			status: http.StatusCreated,
		},
		{
			name:   "list",
			method: http.MethodGet,
			path:   "/api/v1/employees",
			called: "list",
			status: http.StatusOK,
		},
		{
			name:   "get",
			method: http.MethodGet,
			path:   "/api/v1/employees/1",
			called: "get",
			status: http.StatusOK,
		},
		{
			name:   "update",
			method: http.MethodPut,
			path:   "/api/v1/employees/1",
			body:   `{"full_name":"Иван Иванов","department":"ИТ","is_active":false}`,
			called: "update",
			status: http.StatusOK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &employeesServiceStub{
				employee: domain.Employee{
					ID:         1,
					FullName:   "Иван Иванов",
					Department: "ИТ",
					IsActive:   true,
				},
			}
			response := sendEmployeesRequest(testEmployeesRouter(t, service), test.method, test.path, test.body)
			if response.Code != test.status || service.called != test.called {
				t.Fatalf("status = %d, called = %q, body = %s", response.Code, service.called, response.Body.String())
			}
			if response.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("Content-Type = %q", response.Header().Get("Content-Type"))
			}
			if test.name == "get" && service.id != 1 {
				t.Fatalf("service id = %d", service.id)
			}
			if test.name == "update" && (service.id != 1 || service.employee.IsActive) {
				t.Fatalf("update input = %+v, id = %d", service.employee, service.id)
			}
			if test.name != "list" {
				var employee EmployeeResponse
				if err := json.Unmarshal(response.Body.Bytes(), &employee); err != nil {
					t.Fatal(err)
				}
				if employee.ID != 1 || employee.FullName != "Иван Иванов" || employee.Department != "ИТ" || employee.IsActive != (test.name != "update") {
					t.Fatalf("response = %+v", employee)
				}
			}
		})
	}
}

func TestEmployeesEmptyList(t *testing.T) {
	service := &employeesServiceStub{}
	response := sendEmployeesRequest(testEmployeesRouter(t, service), http.MethodGet, "/api/v1/employees", "")
	if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
		t.Fatalf("empty list = %d %s", response.Code, response.Body.String())
	}
}

func TestEmployeeRequestValidation(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "blank full name",
			method: http.MethodPost,
			path:   "/api/v1/employees",
			body:   `{"full_name":"  ","department":"ИТ"}`,
		},
		{
			name:   "long department",
			method: http.MethodPost,
			path:   "/api/v1/employees",
			body:   `{"full_name":"Иван Иванов","department":"` + strings.Repeat("Я", 101) + `"}`,
		},
		{
			name:   "missing active flag",
			method: http.MethodPut,
			path:   "/api/v1/employees/1",
			body:   `{"full_name":"Иван Иванов","department":"ИТ"}`,
		},
		{
			name:   "invalid active flag",
			method: http.MethodPut,
			path:   "/api/v1/employees/1",
			body:   `{"full_name":"Иван Иванов","department":"ИТ","is_active":"false"}`,
		},
		{
			name:   "invalid ID",
			method: http.MethodGet,
			path:   "/api/v1/employees/0",
		},
		{
			name:   "additional JSON",
			method: http.MethodPost,
			path:   "/api/v1/employees",
			body:   `{"full_name":"Иван Иванов","department":"ИТ"} {}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &employeesServiceStub{}
			response := sendEmployeesRequest(testEmployeesRouter(t, service), test.method, test.path, test.body)
			if response.Code != http.StatusBadRequest || service.called != "" || !strings.Contains(response.Body.String(), `"error":"invalid_request"`) {
				t.Fatalf("response = %d %s, service called = %q", response.Code, response.Body.String(), service.called)
			}
		})
	}
}

func TestEmployeeServiceErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{
			name:   "not found",
			err:    core_errors.ErrNotFound,
			status: http.StatusNotFound,
			code:   "not_found",
		},
		{
			name:   "conflict",
			err:    core_errors.ErrConflict,
			status: http.StatusConflict,
			code:   "conflict",
		},
		{
			name:   "internal",
			err:    errors.New("private database detail"),
			status: http.StatusInternalServerError,
			code:   "internal_error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &employeesServiceStub{err: test.err}
			response := sendEmployeesRequest(testEmployeesRouter(t, service), http.MethodGet, "/api/v1/employees/1", "")
			if response.Code != test.status || !strings.Contains(response.Body.String(), `"error":"`+test.code+`"`) || strings.Contains(response.Body.String(), "private database detail") {
				t.Fatalf("response = %d %s", response.Code, response.Body.String())
			}
		})
	}
}
