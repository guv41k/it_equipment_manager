package assignment_transport_http

import (
	"context"
	"net/http"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_http_server "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/server"
)

type AssignmentsService interface {
	IssueEquipment(context.Context, int64, int64) (domain.Assignment, error)
	ReturnEquipment(context.Context, int64) (domain.Assignment, error)
	GetEquipmentAssignments(context.Context, int64) ([]domain.Assignment, error)
}

type AssignmentsHTTPHandler struct {
	assignmentsService AssignmentsService
}

func NewAssignmentsHTTPHandler(service AssignmentsService) *AssignmentsHTTPHandler {
	return &AssignmentsHTTPHandler{
		assignmentsService: service,
	}
}

func (h *AssignmentsHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/equipment/{id}/assignments",
			Handler: h.IssueEquipment,
		},
		{
			Method:  http.MethodPost,
			Path:    "/equipment/{id}/return",
			Handler: h.ReturnEquipment,
		},
		{
			Method:  http.MethodGet,
			Path:    "/equipment/{id}/assignments",
			Handler: h.GetEquipmentAssignments,
		},
	}
}
