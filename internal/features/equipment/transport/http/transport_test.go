package equipment_transport_http

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

type equipmentServiceStub struct {
	called     string
	id         int64
	input      domain.Equipment
	result     domain.Equipment
	list       []domain.Equipment
	categoryID *int64
	status     *domain.EquipmentStatus
	locationID *int64
	limit      *int
	offset     *int
	err        error
}

func (s *equipmentServiceStub) CreateEquipment(_ context.Context, equipment domain.Equipment) (domain.Equipment, error) {
	s.called = "create"
	s.input = equipment
	return s.result, s.err
}

func (s *equipmentServiceStub) GetEquipmentList(_ context.Context, categoryID *int64, status *domain.EquipmentStatus, locationID *int64, limit, offset *int) ([]domain.Equipment, error) {
	s.called = "list"
	s.categoryID = categoryID
	s.status = status
	s.locationID = locationID
	s.limit = limit
	s.offset = offset
	return s.list, s.err
}

func (s *equipmentServiceStub) GetEquipment(_ context.Context, id int64) (domain.Equipment, error) {
	s.called = "get"
	s.id = id
	return s.result, s.err
}

func (s *equipmentServiceStub) UpdateEquipment(_ context.Context, id int64, equipment domain.Equipment) (domain.Equipment, error) {
	s.called = "update"
	s.id = id
	s.input = equipment
	return s.result, s.err
}

func (s *equipmentServiceStub) RetireEquipment(_ context.Context, id int64) (domain.Equipment, error) {
	s.called = "retire"
	s.id = id
	return s.result, s.err
}

func testEquipmentRouter(t *testing.T, service *equipmentServiceStub) http.Handler {
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
	router.AddRoutes(NewEquipmentHTTPHandler(service).Routes()...)
	mux := http.NewServeMux()
	for pattern, handler := range router.Handlers() {
		mux.Handle(pattern, handler)
	}
	return core_http_middleware.Logger(log)(mux)
}

func sendEquipmentRequest(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	handler.ServeHTTP(response, request)
	return response
}

func sampleEquipment() domain.Equipment {
	purchaseDate := time.Date(2025, time.March, 2, 0, 0, 0, 0, time.UTC)
	notes := "В эксплуатации"
	return domain.Equipment{
		ID:              11,
		InventoryNumber: "INV-11",
		SerialNumber:    "SER-11",
		CategoryID:      2,
		Manufacturer:    "Lenovo",
		Model:           "T14",
		Status:          domain.EquipmentStatusInStock,
		LocationID:      3,
		PurchaseDate:    &purchaseDate,
		Notes:           &notes,
		CreatedAt:       time.Date(2025, time.March, 3, 10, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2025, time.March, 3, 10, 0, 0, 0, time.UTC),
	}
}

func TestEquipmentRoutes(t *testing.T) {
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
			path:   "/api/v1/equipment",
			body:   `{"inventory_number":"INV-11","serial_number":"SER-11","category_id":2,"manufacturer":"Lenovo","model":"T14","location_id":3,"purchase_date":"2025-03-02"}`,
			called: "create",
			status: http.StatusCreated,
		},
		{
			name:   "list with filters",
			method: http.MethodGet,
			path:   "/api/v1/equipment?category_id=2&status=in_stock&location_id=3&limit=10&offset=5",
			called: "list",
			status: http.StatusOK,
		},
		{
			name:   "get",
			method: http.MethodGet,
			path:   "/api/v1/equipment/11",
			called: "get",
			status: http.StatusOK,
		},
		{
			name:   "update",
			method: http.MethodPut,
			path:   "/api/v1/equipment/11",
			body:   `{"category_id":2,"manufacturer":"Lenovo","model":"T14","location_id":3,"notes":null}`,
			called: "update",
			status: http.StatusOK,
		},
		{
			name:   "retire",
			method: http.MethodPost,
			path:   "/api/v1/equipment/11/retire",
			called: "retire",
			status: http.StatusOK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			equipment := sampleEquipment()
			if test.name == "retire" {
				equipment.Status = domain.EquipmentStatusRetired
			}
			service := &equipmentServiceStub{
				result: equipment,
				list:   []domain.Equipment{equipment},
			}
			response := sendEquipmentRequest(testEquipmentRouter(t, service), test.method, test.path, test.body)
			if response.Code != test.status || service.called != test.called || response.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("response = %d %s, called = %q", response.Code, response.Body.String(), service.called)
			}
			if test.name == "create" {
				if service.input.InventoryNumber != "INV-11" || service.input.SerialNumber != "SER-11" || service.input.PurchaseDate == nil || service.input.PurchaseDate.Format("2006-01-02") != "2025-03-02" || service.input.Status != "" {
					t.Fatalf("create input = %+v", service.input)
				}
			}
			if test.name == "list with filters" {
				if service.categoryID == nil || *service.categoryID != 2 || service.status == nil || *service.status != domain.EquipmentStatusInStock || service.locationID == nil || *service.locationID != 3 || service.limit == nil || *service.limit != 10 || service.offset == nil || *service.offset != 5 {
					t.Fatalf("list filters = %+v", service)
				}
			}
			if test.name == "get" && service.id != 11 {
				t.Fatalf("get id = %d", service.id)
			}
			if test.name == "update" {
				if service.id != 11 || service.input.InventoryNumber != "" || service.input.Status != "" || service.input.CategoryID != 2 || service.input.Notes != nil {
					t.Fatalf("update input = %+v, id = %d", service.input, service.id)
				}
			}
			if test.name == "retire" && service.id != 11 {
				t.Fatalf("retire id = %d", service.id)
			}
			if test.name != "list with filters" {
				var got EquipmentResponse
				if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				if got.ID != 11 || got.PurchaseDate == nil || *got.PurchaseDate != "2025-03-02" || got.CreatedAt.Format(time.RFC3339) != "2025-03-03T10:00:00Z" {
					t.Fatalf("response = %+v", got)
				}
				if test.name == "retire" && got.Status != domain.EquipmentStatusRetired {
					t.Fatalf("retired response = %+v", got)
				}
			}
		})
	}
}

func TestEquipmentEmptyList(t *testing.T) {
	service := &equipmentServiceStub{}
	response := sendEquipmentRequest(testEquipmentRouter(t, service), http.MethodGet, "/api/v1/equipment", "")
	if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" || service.categoryID != nil || service.limit != nil {
		t.Fatalf("empty list = %d %s", response.Code, response.Body.String())
	}
}

func TestEquipmentRejectsInvalidRequest(t *testing.T) {
	base := `{"inventory_number":"INV-11","serial_number":"SER-11","category_id":2,"manufacturer":"Lenovo","model":"T14","location_id":3}`
	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "blank inventory number",
			method: http.MethodPost,
			path:   "/api/v1/equipment",
			body:   strings.Replace(base, "INV-11", " ", 1),
		},
		{
			name:   "invalid date",
			method: http.MethodPost,
			path:   "/api/v1/equipment",
			body:   strings.TrimSuffix(base, "}") + `,"purchase_date":"2025-02-30"}`,
		},
		{
			name:   "missing category",
			method: http.MethodPost,
			path:   "/api/v1/equipment",
			body:   strings.Replace(base, `"category_id":2,`, "", 1),
		},
		{
			name:   "extra JSON value",
			method: http.MethodPost,
			path:   "/api/v1/equipment",
			body:   base + ` {}`,
		},
		{
			name:   "invalid ID",
			method: http.MethodGet,
			path:   "/api/v1/equipment/0",
		},
		{
			name:   "invalid retirement ID",
			method: http.MethodPost,
			path:   "/api/v1/equipment/0/retire",
		},
		{
			name:   "invalid category filter",
			method: http.MethodGet,
			path:   "/api/v1/equipment?category_id=0",
		},
		{
			name:   "invalid status filter",
			method: http.MethodGet,
			path:   "/api/v1/equipment?status=broken",
		},
		{
			name:   "invalid location filter",
			method: http.MethodGet,
			path:   "/api/v1/equipment?location_id=abc",
		},
		{
			name:   "invalid limit",
			method: http.MethodGet,
			path:   "/api/v1/equipment?limit=0",
		},
		{
			name:   "invalid offset",
			method: http.MethodGet,
			path:   "/api/v1/equipment?offset=-1",
		},
		{
			name:   "invalid update",
			method: http.MethodPut,
			path:   "/api/v1/equipment/11",
			body:   `{"category_id":2,"manufacturer":" ","model":"T14","location_id":3}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &equipmentServiceStub{}
			response := sendEquipmentRequest(testEquipmentRouter(t, service), test.method, test.path, test.body)
			if response.Code != http.StatusBadRequest || service.called != "" || !strings.Contains(response.Body.String(), `"error":"invalid_request"`) {
				t.Fatalf("response = %d %s, service called = %q", response.Code, response.Body.String(), service.called)
			}
		})
	}
}

func TestEquipmentServiceErrors(t *testing.T) {
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
			service := &equipmentServiceStub{err: test.err}
			response := sendEquipmentRequest(testEquipmentRouter(t, service), http.MethodGet, "/api/v1/equipment/11", "")
			if response.Code != test.status || !strings.Contains(response.Body.String(), `"error":"`+test.code+`"`) || strings.Contains(response.Body.String(), "private database detail") {
				t.Fatalf("response = %d %s", response.Code, response.Body.String())
			}
		})
	}
}
