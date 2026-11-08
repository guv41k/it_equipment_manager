package employee_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

func (r *EmployeesRepository) UpdateEmployee(ctx context.Context, employee domain.Employee) (domain.Employee, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Employee{}, fmt.Errorf("begin employee update: %w", err)
	}
	defer tx.Rollback(ctx)

	var lockedID int64
	if err := tx.QueryRow(ctx, "SELECT id FROM employees WHERE id = $1 FOR UPDATE", employee.ID).Scan(&lockedID); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Employee{}, fmt.Errorf("employee id=%d: %w", employee.ID, core_errors.ErrNotFound)
		}
		return domain.Employee{}, fmt.Errorf("lock employee: %w", err)
	}
	if !employee.IsActive {
		query := `
		SELECT EXISTS (
			SELECT 1
			FROM assignments
			WHERE employee_id = $1 AND returned_at IS NULL
		);
		`
		var hasAssignments bool
		if err := tx.QueryRow(ctx, query, employee.ID).Scan(&hasAssignments); err != nil {
			return domain.Employee{}, fmt.Errorf("check active assignments in transaction: %w", err)
		}
		if hasAssignments {
			return domain.Employee{}, fmt.Errorf("employee has active assignments: %w", core_errors.ErrConflict)
		}
	}

	query := `
	UPDATE employees
	SET full_name = $1, department = $2, is_active = $3
	WHERE id = $4
	RETURNING id, full_name, department, is_active;
	`
	var model employeeModel
	if err := model.Scan(tx.QueryRow(ctx, query, employee.FullName, employee.Department, employee.IsActive, employee.ID)); err != nil {
		return domain.Employee{}, fmt.Errorf("update employee: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Employee{}, fmt.Errorf("commit employee update: %w", err)
	}
	return model.toDomain(), nil
}
