package equipment_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

func (r *EquipmentRepository) SaveEquipment(ctx context.Context, equipment domain.Equipment) (domain.Equipment, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	INSERT INTO equipment (
		inventory_number, serial_number, category_id, manufacturer, model,
		status, location_id, purchase_date, warranty_until, notes
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id, inventory_number, serial_number, category_id, manufacturer, model,
		status, location_id, purchase_date, warranty_until, notes, created_at, updated_at;
	`
	var model equipmentModel
	if err := model.Scan(r.pool.QueryRow(
		ctx,
		query,
		equipment.InventoryNumber,
		equipment.SerialNumber,
		equipment.CategoryID,
		equipment.Manufacturer,
		equipment.Model,
		string(equipment.Status),
		equipment.LocationID,
		equipment.PurchaseDate,
		equipment.WarrantyUntil,
		equipment.Notes,
	)); err != nil {
		switch {
		case errors.Is(err, core_postgres_pool.ErrUniqueViolation):
			return domain.Equipment{}, fmt.Errorf("inventory or serial number already exists: %w", core_errors.ErrConflict)
		case errors.Is(err, core_postgres_pool.ErrViolatesForeignKey):
			return domain.Equipment{}, fmt.Errorf("category or location does not exist: %w", core_errors.ErrInvalidArgument)
		default:
			return domain.Equipment{}, fmt.Errorf("insert equipment: %w", err)
		}
	}
	return model.toDomain(), nil
}
