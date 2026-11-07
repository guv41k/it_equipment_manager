package catalog_transport_http

import (
	"net/http"

	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	core_http_response "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/response"
)

func (h *CategoriesHTTPHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	responseHandler := core_http_response.NewHTTPResponseHandler(core_logger.FromContext(r.Context()), w)
	categories, err := h.categoriesService.GetCategories(r.Context())
	if err != nil {
		responseHandler.ErrorResponse(err, "catalog request failed")
		return
	}
	response := make([]CatalogResponse, len(categories))
	for i, category := range categories {
		response[i] = categoryDTOFromDomain(category)
	}
	responseHandler.JSONResponse(response, http.StatusOK)
}
