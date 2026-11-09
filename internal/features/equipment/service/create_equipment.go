package equipment_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

func (s *EquipmentService) CreateEquipment(ctx context.Context, input domain.Equipment) (domain.Equipment, error) {
	equipment := domain.Equipment{
		InventoryNumber: input.InventoryNumber,
		SerialNumber:    input.SerialNumber,
		CategoryID:      input.CategoryID,
		Manufacturer:    input.Manufacturer,
		Model:           input.Model,
		Status:          domain.EquipmentStatusInStock,
		LocationID:      input.LocationID,
		PurchaseDate:    input.PurchaseDate,
		WarrantyUntil:   input.WarrantyUntil,
		Notes:           input.Notes,
	}
	if err := equipment.Validate(); err != nil {
		return domain.Equipment{}, fmt.Errorf("validate equipment: %w", err)
	}
	created, err := s.equipmentRepository.SaveEquipment(ctx, equipment)
	if err != nil {
		return domain.Equipment{}, fmt.Errorf("save equipment in repository: %w", err)
	}
	return created, nil
}
