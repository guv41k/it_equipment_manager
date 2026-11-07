package catalog_transport_http

import (
	"context"
	"net/http"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_http_server "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/server"
)

type CategoriesService interface {
	CreateCategory(context.Context, string, *string) (domain.Category, error)
	GetCategories(context.Context) ([]domain.Category, error)
	GetCategory(context.Context, int64) (domain.Category, error)
	UpdateCategory(context.Context, int64, domain.Category) (domain.Category, error)
}

type CategoriesHTTPHandler struct {
	categoriesService CategoriesService
}

func NewCategoriesHTTPHandler(service CategoriesService) *CategoriesHTTPHandler {
	return &CategoriesHTTPHandler{
		categoriesService: service,
	}
}

func (h *CategoriesHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/equipment-categories",
			Handler: h.CreateCategory,
		},
		{
			Method:  http.MethodGet,
			Path:    "/equipment-categories",
			Handler: h.GetCategories,
		},
		{
			Method:  http.MethodGet,
			Path:    "/equipment-categories/{id}",
			Handler: h.GetCategory,
		},
		{
			Method:  http.MethodPut,
			Path:    "/equipment-categories/{id}",
			Handler: h.UpdateCategory,
		},
	}
}
