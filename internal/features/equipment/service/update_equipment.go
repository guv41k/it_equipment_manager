package equipment_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

func (s *EquipmentService) UpdateEquipment(ctx context.Context, id int64, input domain.Equipment) (domain.Equipment, error) {
	if id <= 0 {
		return domain.Equipment{}, fmt.Errorf("equipment id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	equipment, err := s.equipmentRepository.GetEquipment(ctx, id)
	if err != nil {
		return domain.Equipment{}, fmt.Errorf("get equipment for update: %w", err)
	}
	equipment.CategoryID = input.CategoryID
	equipment.Manufacturer = input.Manufacturer
	equipment.Model = input.Model
	equipment.LocationID = input.LocationID
	equipment.PurchaseDate = input.PurchaseDate
	equipment.WarrantyUntil = input.WarrantyUntil
	equipment.Notes = input.Notes
	if err := equipment.Validate(); err != nil {
		return domain.Equipment{}, fmt.Errorf("validate equipment: %w", err)
	}
	updated, err := s.equipmentRepository.UpdateEquipment(ctx, equipment)
	if err != nil {
		return domain.Equipment{}, fmt.Errorf("update equipment in repository: %w", err)
	}
	return updated, nil
}
