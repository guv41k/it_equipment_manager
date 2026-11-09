package equipment_transport_http

import (
	"net/http"

	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	core_http_request "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/request"
	core_http_response "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/response"
)

func (h *EquipmentHTTPHandler) RetireEquipment(w http.ResponseWriter, r *http.Request) {
	responseHandler := core_http_response.NewHTTPResponseHandler(core_logger.FromContext(r.Context()), w)
	id, err := core_http_request.GetInt64PathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "equipment retirement request failed")
		return
	}
	equipment, err := h.equipmentService.RetireEquipment(r.Context(), id)
	if err != nil {
		responseHandler.ErrorResponse(err, "equipment retirement request failed")
		return
	}
	responseHandler.JSONResponse(equipmentDTOFromDomain(equipment), http.StatusOK)
}
