package equipment_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

func (s *EquipmentService) GetEquipment(ctx context.Context, id int64) (domain.Equipment, error) {
	if id <= 0 {
		return domain.Equipment{}, fmt.Errorf("equipment id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	equipment, err := s.equipmentRepository.GetEquipment(ctx, id)
	if err != nil {
		return domain.Equipment{}, fmt.Errorf("get equipment from repository: %w", err)
	}
	return equipment, nil
}
