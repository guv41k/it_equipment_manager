package employee_postgres_repository

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

func (r *EmployeesRepository) GetEmployees(ctx context.Context) ([]domain.Employee, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, full_name, department, is_active
	FROM employees
	ORDER BY id ASC;
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("select employees: %w", err)
	}
	defer rows.Close()

	employees := make([]domain.Employee, 0)
	for rows.Next() {
		var model employeeModel
		if err := model.Scan(rows); err != nil {
			return nil, fmt.Errorf("scan employee: %w", err)
		}
		employees = append(employees, model.toDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate employees: %w", err)
	}
	return employees, nil
}
