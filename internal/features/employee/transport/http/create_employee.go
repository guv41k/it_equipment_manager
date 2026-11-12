package employee_transport_http

import (
	"net/http"

	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	core_http_request "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/request"
	core_http_response "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/response"
)

func (h *EmployeesHTTPHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	responseHandler := core_http_response.NewHTTPResponseHandler(core_logger.FromContext(r.Context()), w)
	var request CreateEmployeeRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "employee request failed")
		return
	}
	employee, err := h.employeesService.CreateEmployee(r.Context(), request.FullName, request.Department)
	if err != nil {
		responseHandler.ErrorResponse(err, "employee request failed")
		return
	}
	responseHandler.JSONResponse(employeeDTOFromDomain(employee), http.StatusCreated)
}
