package employee_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

func (s *EmployeesService) UpdateEmployee(ctx context.Context, id int64, employee domain.Employee) (domain.Employee, error) {
	if id <= 0 {
		return domain.Employee{}, fmt.Errorf("employee id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	employee.ID = id
	if err := employee.Validate(); err != nil {
		return domain.Employee{}, fmt.Errorf("validate employee: %w", err)
	}
	if !employee.IsActive {
		hasAssignments, err := s.employeesRepository.HasActiveAssignments(ctx, id)
		if err != nil {
			return domain.Employee{}, fmt.Errorf("check employee assignments: %w", err)
		}
		if hasAssignments {
			return domain.Employee{}, fmt.Errorf("employee has active assignments: %w", core_errors.ErrConflict)
		}
	}
	updated, err := s.employeesRepository.UpdateEmployee(ctx, employee)
	if err != nil {
		return domain.Employee{}, fmt.Errorf("update employee in repository: %w", err)
	}
	return updated, nil
}
