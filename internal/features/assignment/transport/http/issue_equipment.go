package assignment_transport_http

import (
	"net/http"

	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	core_http_request "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/request"
	core_http_response "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/response"
)

func (h *AssignmentsHTTPHandler) IssueEquipment(w http.ResponseWriter, r *http.Request) {
	responseHandler := core_http_response.NewHTTPResponseHandler(core_logger.FromContext(r.Context()), w)
	equipmentID, err := core_http_request.GetInt64PathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "assignment request failed")
		return
	}
	var request IssueEquipmentRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "assignment request failed")
		return
	}
	assignment, err := h.assignmentsService.IssueEquipment(r.Context(), equipmentID, request.EmployeeID)
	if err != nil {
		responseHandler.ErrorResponse(err, "assignment request failed")
		return
	}
	responseHandler.JSONResponse(assignmentDTOFromDomain(assignment), http.StatusCreated)
}
