package equipment_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

func (r *EquipmentRepository) UpdateEquipment(ctx context.Context, equipment domain.Equipment) (domain.Equipment, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE equipment
	SET category_id = $1,
		manufacturer = $2,
		model = $3,
		location_id = $4,
		purchase_date = $5,
		warranty_until = $6,
		notes = $7,
		updated_at = NOW()
	WHERE id = $8
	RETURNING id, inventory_number, serial_number, category_id, manufacturer, model,
		status, location_id, purchase_date, warranty_until, notes, created_at, updated_at;
	`
	var model equipmentModel
	if err := model.Scan(r.pool.QueryRow(
		ctx,
		query,
		equipment.CategoryID,
		equipment.Manufacturer,
		equipment.Model,
		equipment.LocationID,
		equipment.PurchaseDate,
		equipment.WarrantyUntil,
		equipment.Notes,
		equipment.ID,
	)); err != nil {
		switch {
		case errors.Is(err, core_postgres_pool.ErrNoRows):
			return domain.Equipment{}, fmt.Errorf("equipment id=%d: %w", equipment.ID, core_errors.ErrNotFound)
		case errors.Is(err, core_postgres_pool.ErrViolatesForeignKey):
			return domain.Equipment{}, fmt.Errorf("category or location does not exist: %w", core_errors.ErrInvalidArgument)
		default:
			return domain.Equipment{}, fmt.Errorf("update equipment: %w", err)
		}
	}
	return model.toDomain(), nil
}
