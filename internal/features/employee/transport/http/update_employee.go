package employee_transport_http

import (
	"net/http"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	core_http_request "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/request"
	core_http_response "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/response"
)

func (h *EmployeesHTTPHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	responseHandler := core_http_response.NewHTTPResponseHandler(core_logger.FromContext(r.Context()), w)
	id, err := core_http_request.GetInt64PathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "employee request failed")
		return
	}
	var request UpdateEmployeeRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "employee request failed")
		return
	}
	employee, err := h.employeesService.UpdateEmployee(r.Context(), id, domain.Employee{
		FullName:   request.FullName,
		Department: request.Department,
		IsActive:   *request.IsActive,
	})
	if err != nil {
		responseHandler.ErrorResponse(err, "employee request failed")
		return
	}
	responseHandler.JSONResponse(employeeDTOFromDomain(employee), http.StatusOK)
}
