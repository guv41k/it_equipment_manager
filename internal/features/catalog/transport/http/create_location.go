package catalog_transport_http

import (
	"net/http"

	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	core_http_request "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/request"
	core_http_response "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/response"
)

func (h *LocationsHTTPHandler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	responseHandler := core_http_response.NewHTTPResponseHandler(core_logger.FromContext(r.Context()), w)
	var request CatalogRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "catalog request failed")
		return
	}
	location, err := h.locationsService.CreateLocation(r.Context(), request.Name, request.Description)
	if err != nil {
		responseHandler.ErrorResponse(err, "catalog request failed")
		return
	}
	responseHandler.JSONResponse(locationDTOFromDomain(location), http.StatusCreated)
}
