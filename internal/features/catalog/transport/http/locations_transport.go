package catalog_transport_http

import (
	"context"
	"net/http"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_http_server "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/server"
)

type LocationsService interface {
	CreateLocation(context.Context, string, *string) (domain.Location, error)
	GetLocations(context.Context) ([]domain.Location, error)
	GetLocation(context.Context, int64) (domain.Location, error)
	UpdateLocation(context.Context, int64, domain.Location) (domain.Location, error)
}

type LocationsHTTPHandler struct {
	locationsService LocationsService
}

func NewLocationsHTTPHandler(service LocationsService) *LocationsHTTPHandler {
	return &LocationsHTTPHandler{
		locationsService: service,
	}
}

func (h *LocationsHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/locations",
			Handler: h.CreateLocation,
		},
		{
			Method:  http.MethodGet,
			Path:    "/locations",
			Handler: h.GetLocations,
		},
		{
			Method:  http.MethodGet,
			Path:    "/locations/{id}",
			Handler: h.GetLocation,
		},
		{
			Method:  http.MethodPut,
			Path:    "/locations/{id}",
			Handler: h.UpdateLocation,
		},
	}
}
