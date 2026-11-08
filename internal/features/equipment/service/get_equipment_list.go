package equipment_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

func (s *EquipmentService) GetEquipmentList(ctx context.Context, categoryID *int64, status *domain.EquipmentStatus, locationID *int64, limit, offset *int) ([]domain.Equipment, error) {
	if categoryID != nil && *categoryID <= 0 {
		return nil, fmt.Errorf("category id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	if status != nil {
		switch *status {
		case domain.EquipmentStatusInStock, domain.EquipmentStatusAssigned, domain.EquipmentStatusRetired:
		default:
			return nil, fmt.Errorf("invalid equipment status: %w", core_errors.ErrInvalidArgument)
		}
	}
	if locationID != nil && *locationID <= 0 {
		return nil, fmt.Errorf("location id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	if limit != nil && *limit <= 0 {
		return nil, fmt.Errorf("limit must be positive: %w", core_errors.ErrInvalidArgument)
	}
	if offset != nil && *offset < 0 {
		return nil, fmt.Errorf("offset must not be negative: %w", core_errors.ErrInvalidArgument)
	}
	equipment, err := s.equipmentRepository.GetEquipmentList(ctx, categoryID, status, locationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get equipment list from repository: %w", err)
	}
	return equipment, nil
}
