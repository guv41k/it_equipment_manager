package catalog_transport_http

import (
	"net/http"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	core_http_request "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/request"
	core_http_response "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/response"
)

func (h *CategoriesHTTPHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	responseHandler := core_http_response.NewHTTPResponseHandler(core_logger.FromContext(r.Context()), w)
	id, err := core_http_request.GetInt64PathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "catalog request failed")
		return
	}
	var request CatalogRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "catalog request failed")
		return
	}
	category, err := h.categoriesService.UpdateCategory(r.Context(), id, domain.Category{
		Name:        request.Name,
		Description: request.Description,
	})
	if err != nil {
		responseHandler.ErrorResponse(err, "catalog request failed")
		return
	}
	responseHandler.JSONResponse(categoryDTOFromDomain(category), http.StatusOK)
}
