package employee_transport_http

import (
	"context"
	"net/http"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_http_server "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/server"
)

type EmployeesService interface {
	CreateEmployee(context.Context, string, string) (domain.Employee, error)
	GetEmployees(context.Context) ([]domain.Employee, error)
	GetEmployee(context.Context, int64) (domain.Employee, error)
	UpdateEmployee(context.Context, int64, domain.Employee) (domain.Employee, error)
}

type EmployeesHTTPHandler struct {
	employeesService EmployeesService
}

func NewEmployeesHTTPHandler(service EmployeesService) *EmployeesHTTPHandler {
	return &EmployeesHTTPHandler{
		employeesService: service,
	}
}

func (h *EmployeesHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/employees",
			Handler: h.CreateEmployee,
		},
		{
			Method:  http.MethodGet,
			Path:    "/employees",
			Handler: h.GetEmployees,
		},
		{
			Method:  http.MethodGet,
			Path:    "/employees/{id}",
			Handler: h.GetEmployee,
		},
		{
			Method:  http.MethodPut,
			Path:    "/employees/{id}",
			Handler: h.UpdateEmployee,
		},
	}
}
