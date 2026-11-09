package assignment_transport_http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	core_http_middleware "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/middleware"
	core_http_server "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/server"
)

type assignmentsServiceStub struct {
	called      string
	equipmentID int64
	employeeID  int64
	assignment  domain.Assignment
	history     []domain.Assignment
	err         error
}

func (s *assignmentsServiceStub) IssueEquipment(_ context.Context, equipmentID, employeeID int64) (domain.Assignment, error) {
	s.called = "issue"
	s.equipmentID = equipmentID
	s.employeeID = employeeID
	return s.assignment, s.err
}

func (s *assignmentsServiceStub) ReturnEquipment(_ context.Context, equipmentID int64) (domain.Assignment, error) {
	s.called = "return"
	s.equipmentID = equipmentID
	return s.assignment, s.err
}

func (s *assignmentsServiceStub) GetEquipmentAssignments(_ context.Context, equipmentID int64) ([]domain.Assignment, error) {
	s.called = "history"
	s.equipmentID = equipmentID
	return s.history, s.err
}

func testAssignmentsRouter(t *testing.T, service *assignmentsServiceStub) http.Handler {
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
	router.AddRoutes(NewAssignmentsHTTPHandler(service).Routes()...)
	mux := http.NewServeMux()
	for pattern, handler := range router.Handlers() {
		mux.Handle(pattern, handler)
	}
	return core_http_middleware.Logger(log)(mux)
}

func sendAssignmentsRequest(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	handler.ServeHTTP(response, request)
	return response
}

func TestAssignmentRoutes(t *testing.T) {
	issuedAt := time.Date(2026, time.November, 9, 10, 0, 0, 0, time.UTC)
	returnedAt := issuedAt.Add(time.Hour)
	issuedAssignment := domain.Assignment{
		ID:          3,
		EquipmentID: 11,
		EmployeeID:  7,
		IssuedAt:    issuedAt,
	}
	returnedAssignment := issuedAssignment
	returnedAssignment.ReturnedAt = &returnedAt
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		called string
		status int
	}{
		{
			name:   "issue",
			method: http.MethodPost,
			path:   "/api/v1/equipment/11/assignments",
			body:   `{"employee_id":7}`,
			called: "issue",
			status: http.StatusCreated,
		},
		{
			name:   "return",
			method: http.MethodPost,
			path:   "/api/v1/equipment/11/return",
			called: "return",
			status: http.StatusOK,
		},
		{
			name:   "history",
			method: http.MethodGet,
			path:   "/api/v1/equipment/11/assignments",
			called: "history",
			status: http.StatusOK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &assignmentsServiceStub{
				assignment: issuedAssignment,
				history:    []domain.Assignment{returnedAssignment},
			}
			if test.name == "return" {
				service.assignment = returnedAssignment
			}
			response := sendAssignmentsRequest(testAssignmentsRouter(t, service), test.method, test.path, test.body)
			if response.Code != test.status || service.called != test.called || service.equipmentID != 11 || response.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("response = %d %s, service = %+v", response.Code, response.Body.String(), service)
			}
			if test.name == "issue" && service.employeeID != 7 {
				t.Fatalf("employee id = %d", service.employeeID)
			}
			var got []AssignmentResponse
			if test.name == "history" {
				if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
			} else {
				var item AssignmentResponse
				if err := json.Unmarshal(response.Body.Bytes(), &item); err != nil {
					t.Fatal(err)
				}
				got = []AssignmentResponse{item}
			}
			if len(got) != 1 || got[0].ID != 3 || got[0].EquipmentID != 11 || got[0].EmployeeID != 7 || !got[0].IssuedAt.Equal(issuedAt) {
				t.Fatalf("response = %+v", got)
			}
			if test.name == "issue" {
				if got[0].ReturnedAt != nil || !strings.Contains(response.Body.String(), `"returned_at":null`) {
					t.Fatalf("issue response = %s", response.Body.String())
				}
			} else if got[0].ReturnedAt == nil || !got[0].ReturnedAt.Equal(returnedAt) {
				t.Fatalf("returned_at = %v", got[0].ReturnedAt)
			}
		})
	}
}

func TestAssignmentEmptyHistory(t *testing.T) {
	service := &assignmentsServiceStub{}
	response := sendAssignmentsRequest(testAssignmentsRouter(t, service), http.MethodGet, "/api/v1/equipment/11/assignments", "")
	if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
		t.Fatalf("empty history = %d %s", response.Code, response.Body.String())
	}
}

func TestAssignmentInvalidRequest(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "invalid equipment id",
			method: http.MethodGet,
			path:   "/api/v1/equipment/0/assignments",
		},
		{
			name:   "missing employee id",
			method: http.MethodPost,
			path:   "/api/v1/equipment/11/assignments",
			body:   `{}`,
		},
		{
			name:   "invalid employee id",
			method: http.MethodPost,
			path:   "/api/v1/equipment/11/assignments",
			body:   `{"employee_id":-1}`,
		},
		{
			name:   "invalid employee type",
			method: http.MethodPost,
			path:   "/api/v1/equipment/11/assignments",
			body:   `{"employee_id":"7"}`,
		},
		{
			name:   "extra JSON value",
			method: http.MethodPost,
			path:   "/api/v1/equipment/11/assignments",
			body:   `{"employee_id":7} {}`,
		},
		{
			name:   "invalid return id",
			method: http.MethodPost,
			path:   "/api/v1/equipment/0/return",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &assignmentsServiceStub{}
			response := sendAssignmentsRequest(testAssignmentsRouter(t, service), test.method, test.path, test.body)
			if response.Code != http.StatusBadRequest || service.called != "" || !strings.Contains(response.Body.String(), `"error":"invalid_request"`) {
				t.Fatalf("response = %d %s, service called = %q", response.Code, response.Body.String(), service.called)
			}
		})
	}
}

func TestAssignmentServiceErrors(t *testing.T) {
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
			service := &assignmentsServiceStub{err: test.err}
			response := sendAssignmentsRequest(testAssignmentsRouter(t, service), http.MethodPost, "/api/v1/equipment/11/return", "")
			if response.Code != test.status || !strings.Contains(response.Body.String(), `"error":"`+test.code+`"`) || strings.Contains(response.Body.String(), "private database detail") {
				t.Fatalf("response = %d %s", response.Code, response.Body.String())
			}
			if test.name == "conflict" && !strings.Contains(response.Body.String(), `"message":"operation conflicts with current state"`) {
				t.Fatalf("conflict response = %s", response.Body.String())
			}
		})
	}
}
