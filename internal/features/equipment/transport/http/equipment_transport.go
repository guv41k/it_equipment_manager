package equipment_transport_http

import (
	"context"
	"net/http"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_http_server "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/server"
)

type EquipmentService interface {
	CreateEquipment(context.Context, domain.Equipment) (domain.Equipment, error)
	GetEquipmentList(context.Context, *int64, *domain.EquipmentStatus, *int64, *int, *int) ([]domain.Equipment, error)
	GetEquipment(context.Context, int64) (domain.Equipment, error)
	UpdateEquipment(context.Context, int64, domain.Equipment) (domain.Equipment, error)
}

type EquipmentHTTPHandler struct {
	equipmentService EquipmentService
}

func NewEquipmentHTTPHandler(service EquipmentService) *EquipmentHTTPHandler {
	return &EquipmentHTTPHandler{
		equipmentService: service,
	}
}

func (h *EquipmentHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/equipment",
			Handler: h.CreateEquipment,
		},
		{
			Method:  http.MethodGet,
			Path:    "/equipment",
			Handler: h.GetEquipmentList,
		},
		{
			Method:  http.MethodGet,
			Path:    "/equipment/{id}",
			Handler: h.GetEquipment,
		},
		{
			Method:  http.MethodPut,
			Path:    "/equipment/{id}",
			Handler: h.UpdateEquipment,
		},
	}
}
