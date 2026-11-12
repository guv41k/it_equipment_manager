package assignment_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

func (r *AssignmentsRepository) SaveAssignment(ctx context.Context, assignment domain.Assignment) (domain.Assignment, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Assignment{}, fmt.Errorf("begin assignment issue: %w", err)
	}
	defer tx.Rollback(ctx)

	var status string
	if err := tx.QueryRow(ctx, "SELECT status FROM equipment WHERE id = $1 FOR UPDATE", assignment.EquipmentID).Scan(&status); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Assignment{}, fmt.Errorf("equipment id=%d: %w", assignment.EquipmentID, core_errors.ErrNotFound)
		}
		return domain.Assignment{}, fmt.Errorf("lock equipment for issue: %w", err)
	}
	if domain.EquipmentStatus(status) != domain.EquipmentStatusInStock {
		return domain.Assignment{}, fmt.Errorf("equipment is not in stock: %w", core_errors.ErrConflict)
	}

	var isActive bool
	if err := tx.QueryRow(ctx, "SELECT is_active FROM employees WHERE id = $1 FOR UPDATE", assignment.EmployeeID).Scan(&isActive); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Assignment{}, fmt.Errorf("employee id=%d: %w", assignment.EmployeeID, core_errors.ErrNotFound)
		}
		return domain.Assignment{}, fmt.Errorf("lock employee for issue: %w", err)
	}
	if !isActive {
		return domain.Assignment{}, fmt.Errorf("employee is inactive: %w", core_errors.ErrConflict)
	}

	var hasActive bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM assignments
			WHERE equipment_id = $1 AND returned_at IS NULL
		)
	`, assignment.EquipmentID).Scan(&hasActive); err != nil {
		return domain.Assignment{}, fmt.Errorf("check active assignment in transaction: %w", err)
	}
	if hasActive {
		return domain.Assignment{}, fmt.Errorf("equipment has active assignment: %w", core_errors.ErrConflict)
	}

	var model assignmentModel
	if err := model.Scan(tx.QueryRow(ctx, `
		INSERT INTO assignments (equipment_id, employee_id, issued_at)
		VALUES ($1, $2, clock_timestamp())
		RETURNING id, equipment_id, employee_id, issued_at, returned_at
	`, assignment.EquipmentID, assignment.EmployeeID)); err != nil {
		if errors.Is(err, core_postgres_pool.ErrViolatesForeignKey) {
			return domain.Assignment{}, fmt.Errorf("equipment or employee does not exist: %w", core_errors.ErrNotFound)
		}
		return domain.Assignment{}, fmt.Errorf("insert assignment: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE equipment
		SET status = 'assigned', updated_at = GREATEST(clock_timestamp(), $2::timestamptz)
		WHERE id = $1
	`, assignment.EquipmentID, model.IssuedAt); err != nil {
		return domain.Assignment{}, fmt.Errorf("mark equipment assigned: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Assignment{}, fmt.Errorf("commit assignment issue: %w", err)
	}
	return model.toDomain(), nil
}
