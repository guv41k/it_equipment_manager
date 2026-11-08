package equipment_transport_http

import (
	"fmt"
	"net/http"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	core_http_request "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/request"
	core_http_response "gitlab.com/guv41k/it_equipment_manager/internal/core/transport/http/response"
)

type equipmentListQuery struct {
	categoryID *int64
	status     *domain.EquipmentStatus
	locationID *int64
	limit      *int
	offset     *int
}

func (h *EquipmentHTTPHandler) GetEquipmentList(w http.ResponseWriter, r *http.Request) {
	responseHandler := core_http_response.NewHTTPResponseHandler(core_logger.FromContext(r.Context()), w)
	query, err := parseEquipmentListQuery(r)
	if err != nil {
		responseHandler.ErrorResponse(err, "equipment request failed")
		return
	}
	equipment, err := h.equipmentService.GetEquipmentList(r.Context(), query.categoryID, query.status, query.locationID, query.limit, query.offset)
	if err != nil {
		responseHandler.ErrorResponse(err, "equipment request failed")
		return
	}
	response := make([]EquipmentResponse, len(equipment))
	for i, item := range equipment {
		response[i] = equipmentDTOFromDomain(item)
	}
	responseHandler.JSONResponse(response, http.StatusOK)
}

func parseEquipmentListQuery(r *http.Request) (equipmentListQuery, error) {
	var query equipmentListQuery
	var err error
	query.categoryID, err = core_http_request.GetInt64QueryParam(r, "category_id")
	if err != nil {
		return equipmentListQuery{}, err
	}
	if query.categoryID != nil && *query.categoryID <= 0 {
		return equipmentListQuery{}, fmt.Errorf("category_id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	status := domain.EquipmentStatus(r.URL.Query().Get("status"))
	if status != "" {
		switch status {
		case domain.EquipmentStatusInStock, domain.EquipmentStatusAssigned, domain.EquipmentStatusRetired:
			query.status = &status
		default:
			return equipmentListQuery{}, fmt.Errorf("invalid equipment status: %w", core_errors.ErrInvalidArgument)
		}
	}
	query.locationID, err = core_http_request.GetInt64QueryParam(r, "location_id")
	if err != nil {
		return equipmentListQuery{}, err
	}
	if query.locationID != nil && *query.locationID <= 0 {
		return equipmentListQuery{}, fmt.Errorf("location_id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	query.limit, err = core_http_request.GetIntQueryParam(r, "limit")
	if err != nil {
		return equipmentListQuery{}, err
	}
	if query.limit != nil && *query.limit <= 0 {
		return equipmentListQuery{}, fmt.Errorf("limit must be positive: %w", core_errors.ErrInvalidArgument)
	}
	query.offset, err = core_http_request.GetIntQueryParam(r, "offset")
	if err != nil {
		return equipmentListQuery{}, err
	}
	if query.offset != nil && *query.offset < 0 {
		return equipmentListQuery{}, fmt.Errorf("offset must not be negative: %w", core_errors.ErrInvalidArgument)
	}
	return query, nil
}
