package equipment_transport_http

import (
	"net/http"

	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	core_http_request "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/request"
	core_http_response "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/response"
)

func (h *EquipmentHTTPHandler) UpdateEquipment(w http.ResponseWriter, r *http.Request) {
	responseHandler := core_http_response.NewHTTPResponseHandler(core_logger.FromContext(r.Context()), w)
	id, err := core_http_request.GetInt64PathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "equipment request failed")
		return
	}
	var request UpdateEquipmentRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "equipment request failed")
		return
	}
	input, err := equipmentFromUpdateRequest(request)
	if err != nil {
		responseHandler.ErrorResponse(err, "equipment request failed")
		return
	}
	equipment, err := h.equipmentService.UpdateEquipment(r.Context(), id, input)
	if err != nil {
		responseHandler.ErrorResponse(err, "equipment request failed")
		return
	}
	responseHandler.JSONResponse(equipmentDTOFromDomain(equipment), http.StatusOK)
}
