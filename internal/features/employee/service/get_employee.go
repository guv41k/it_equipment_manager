package employee_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

func (s *EmployeesService) GetEmployee(ctx context.Context, id int64) (domain.Employee, error) {
	if id <= 0 {
		return domain.Employee{}, fmt.Errorf("employee id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	employee, err := s.employeesRepository.GetEmployee(ctx, id)
	if err != nil {
		return domain.Employee{}, fmt.Errorf("get employee from repository: %w", err)
	}
	return employee, nil
}
