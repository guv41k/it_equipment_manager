package employee_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

func (s *EmployeesService) CreateEmployee(ctx context.Context, fullName, department string) (domain.Employee, error) {
	employee := domain.Employee{
		FullName:   fullName,
		Department: department,
		IsActive:   true,
	}
	if err := employee.Validate(); err != nil {
		return domain.Employee{}, fmt.Errorf("validate employee: %w", err)
	}
	employee, err := s.employeesRepository.SaveEmployee(ctx, employee)
	if err != nil {
		return domain.Employee{}, fmt.Errorf("save employee in repository: %w", err)
	}
	return employee, nil
}
