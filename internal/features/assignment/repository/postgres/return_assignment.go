package assignment_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

func (r *AssignmentsRepository) ReturnAssignment(ctx context.Context, equipmentID int64) (domain.Assignment, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Assignment{}, fmt.Errorf("begin assignment return: %w", err)
	}
	defer tx.Rollback(ctx)

	var status string
	if err := tx.QueryRow(ctx, "SELECT status FROM equipment WHERE id = $1 FOR UPDATE", equipmentID).Scan(&status); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Assignment{}, fmt.Errorf("equipment id=%d: %w", equipmentID, core_errors.ErrNotFound)
		}
		return domain.Assignment{}, fmt.Errorf("lock equipment for return: %w", err)
	}
	if domain.EquipmentStatus(status) != domain.EquipmentStatusAssigned {
		return domain.Assignment{}, fmt.Errorf("equipment is not assigned: %w", core_errors.ErrConflict)
	}

	var model assignmentModel
	if err := model.Scan(tx.QueryRow(ctx, `
		UPDATE assignments
		SET returned_at = GREATEST(clock_timestamp(), issued_at)
		WHERE equipment_id = $1 AND returned_at IS NULL
		RETURNING id, equipment_id, employee_id, issued_at, returned_at
	`, equipmentID)); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Assignment{}, fmt.Errorf("equipment has no active assignment: %w", core_errors.ErrConflict)
		}
		return domain.Assignment{}, fmt.Errorf("close assignment: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE equipment
		SET status = 'in_stock', updated_at = GREATEST(clock_timestamp(), $2::timestamptz)
		WHERE id = $1
	`, equipmentID, model.ReturnedAt); err != nil {
		return domain.Assignment{}, fmt.Errorf("mark equipment in stock: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Assignment{}, fmt.Errorf("commit assignment return: %w", err)
	}
	return model.toDomain(), nil
}
