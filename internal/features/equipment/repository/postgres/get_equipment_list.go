package equipment_postgres_repository

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

func (r *EquipmentRepository) GetEquipmentList(ctx context.Context, categoryID *int64, status *domain.EquipmentStatus, locationID *int64, limit, offset *int) ([]domain.Equipment, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var statusValue any
	if status != nil {
		statusValue = string(*status)
	}
	query := `
	SELECT id, inventory_number, serial_number, category_id, manufacturer, model,
		status, location_id, purchase_date, warranty_until, notes, created_at, updated_at
	FROM equipment
	WHERE ($1::BIGINT IS NULL OR category_id = $1)
		AND ($2::VARCHAR IS NULL OR status = $2)
		AND ($3::BIGINT IS NULL OR location_id = $3)
	ORDER BY id ASC
	LIMIT $4 OFFSET $5;
	`
	rows, err := r.pool.Query(ctx, query, categoryID, statusValue, locationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("select equipment list: %w", err)
	}
	defer rows.Close()

	equipment := make([]domain.Equipment, 0)
	for rows.Next() {
		var model equipmentModel
		if err := model.Scan(rows); err != nil {
			return nil, fmt.Errorf("scan equipment: %w", err)
		}
		equipment = append(equipment, model.toDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate equipment list: %w", err)
	}
	return equipment, nil
}
