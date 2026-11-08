package employee_postgres_repository

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

func (r *EmployeesRepository) SaveEmployee(ctx context.Context, employee domain.Employee) (domain.Employee, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	INSERT INTO employees (full_name, department, is_active)
	VALUES ($1, $2, $3)
	RETURNING id, full_name, department, is_active;
	`
	var model employeeModel
	if err := model.Scan(r.pool.QueryRow(ctx, query, employee.FullName, employee.Department, employee.IsActive)); err != nil {
		return domain.Employee{}, fmt.Errorf("insert employee: %w", err)
	}
	return model.toDomain(), nil
}
