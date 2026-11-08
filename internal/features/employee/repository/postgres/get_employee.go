package employee_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

func (r *EmployeesRepository) GetEmployee(ctx context.Context, id int64) (domain.Employee, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, full_name, department, is_active
	FROM employees
	WHERE id = $1;
	`
	var model employeeModel
	if err := model.Scan(r.pool.QueryRow(ctx, query, id)); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Employee{}, fmt.Errorf("employee id=%d: %w", id, core_errors.ErrNotFound)
		}
		return domain.Employee{}, fmt.Errorf("select employee: %w", err)
	}
	return model.toDomain(), nil
}
