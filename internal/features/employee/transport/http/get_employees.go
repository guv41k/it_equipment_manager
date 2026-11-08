package employee_transport_http

import (
	"net/http"

	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	core_http_response "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/response"
)

func (h *EmployeesHTTPHandler) GetEmployees(w http.ResponseWriter, r *http.Request) {
	responseHandler := core_http_response.NewHTTPResponseHandler(core_logger.FromContext(r.Context()), w)
	employees, err := h.employeesService.GetEmployees(r.Context())
	if err != nil {
		responseHandler.ErrorResponse(err, "employee request failed")
		return
	}
	response := make([]EmployeeResponse, len(employees))
	for i, employee := range employees {
		response[i] = employeeDTOFromDomain(employee)
	}
	responseHandler.JSONResponse(response, http.StatusOK)
}
