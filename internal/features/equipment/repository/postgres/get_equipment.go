package equipment_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

func (r *EquipmentRepository) GetEquipment(ctx context.Context, id int64) (domain.Equipment, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, inventory_number, serial_number, category_id, manufacturer, model,
		status, location_id, purchase_date, warranty_until, notes, created_at, updated_at
	FROM equipment
	WHERE id = $1;
	`
	var model equipmentModel
	if err := model.Scan(r.pool.QueryRow(ctx, query, id)); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Equipment{}, fmt.Errorf("equipment id=%d: %w", id, core_errors.ErrNotFound)
		}
		return domain.Equipment{}, fmt.Errorf("select equipment: %w", err)
	}
	return model.toDomain(), nil
}
