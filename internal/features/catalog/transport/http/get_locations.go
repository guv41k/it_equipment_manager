package catalog_transport_http

import (
	"net/http"

	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	core_http_response "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/response"
)

func (h *LocationsHTTPHandler) GetLocations(w http.ResponseWriter, r *http.Request) {
	responseHandler := core_http_response.NewHTTPResponseHandler(core_logger.FromContext(r.Context()), w)
	locations, err := h.locationsService.GetLocations(r.Context())
	if err != nil {
		responseHandler.ErrorResponse(err, "catalog request failed")
		return
	}
	response := make([]CatalogResponse, len(locations))
	for i, location := range locations {
		response[i] = locationDTOFromDomain(location)
	}
	responseHandler.JSONResponse(response, http.StatusOK)
}
