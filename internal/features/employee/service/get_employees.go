package employee_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

func (s *EmployeesService) GetEmployees(ctx context.Context) ([]domain.Employee, error) {
	employees, err := s.employeesRepository.GetEmployees(ctx)
	if err != nil {
		return nil, fmt.Errorf("get employees from repository: %w", err)
	}
	return employees, nil
}
