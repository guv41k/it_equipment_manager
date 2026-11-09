package equipment_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

func (r *EquipmentRepository) RetireEquipment(ctx context.Context, id int64) (domain.Equipment, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Equipment{}, fmt.Errorf("begin equipment retirement: %w", err)
	}
	defer tx.Rollback(ctx)

	var status string
	if err := tx.QueryRow(ctx, "SELECT status FROM equipment WHERE id = $1 FOR UPDATE", id).Scan(&status); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Equipment{}, fmt.Errorf("equipment id=%d: %w", id, core_errors.ErrNotFound)
		}
		return domain.Equipment{}, fmt.Errorf("lock equipment for retirement: %w", err)
	}
	if domain.EquipmentStatus(status) != domain.EquipmentStatusInStock {
		return domain.Equipment{}, fmt.Errorf("equipment is not in stock: %w", core_errors.ErrConflict)
	}

	var hasActive bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM assignments
			WHERE equipment_id = $1 AND returned_at IS NULL
		)
	`, id).Scan(&hasActive); err != nil {
		return domain.Equipment{}, fmt.Errorf("check active assignment for retirement: %w", err)
	}
	if hasActive {
		return domain.Equipment{}, fmt.Errorf("equipment has active assignment: %w", core_errors.ErrConflict)
	}

	var model equipmentModel
	if err := model.Scan(tx.QueryRow(ctx, `
		UPDATE equipment
		SET status = 'retired', updated_at = clock_timestamp()
		WHERE id = $1
		RETURNING id, inventory_number, serial_number, category_id, manufacturer, model,
			status, location_id, purchase_date, warranty_until, notes, created_at, updated_at
	`, id)); err != nil {
		return domain.Equipment{}, fmt.Errorf("update retired equipment: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Equipment{}, fmt.Errorf("commit equipment retirement: %w", err)
	}
	return model.toDomain(), nil
}
