package employee_service

import (
	"context"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

type EmployeesRepository interface {
	SaveEmployee(context.Context, domain.Employee) (domain.Employee, error)
	GetEmployees(context.Context) ([]domain.Employee, error)
	GetEmployee(context.Context, int64) (domain.Employee, error)
	HasActiveAssignments(context.Context, int64) (bool, error)
	UpdateEmployee(context.Context, domain.Employee) (domain.Employee, error)
}

type EmployeesService struct {
	employeesRepository EmployeesRepository
}

func NewEmployeesService(repository EmployeesRepository) *EmployeesService {
	return &EmployeesService{
		employeesRepository: repository,
	}
}
