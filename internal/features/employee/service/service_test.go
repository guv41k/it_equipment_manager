package employee_service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

type employeesRepositoryStub struct {
	ctx                  context.Context
	id                   int64
	employee             domain.Employee
	result               domain.Employee
	list                 []domain.Employee
	hasActiveAssignments bool
	saveCalls            int
	listCalls            int
	getCalls             int
	assignmentCheckCalls int
	updateCalls          int
	err                  error
}

func (r *employeesRepositoryStub) SaveEmployee(ctx context.Context, employee domain.Employee) (domain.Employee, error) {
	r.ctx = ctx
	r.employee = employee
	r.saveCalls++
	return r.result, r.err
}

func (r *employeesRepositoryStub) GetEmployees(ctx context.Context) ([]domain.Employee, error) {
	r.ctx = ctx
	r.listCalls++
	return r.list, r.err
}

func (r *employeesRepositoryStub) GetEmployee(ctx context.Context, id int64) (domain.Employee, error) {
	r.ctx = ctx
	r.id = id
	r.getCalls++
	return r.result, r.err
}

func (r *employeesRepositoryStub) HasActiveAssignments(ctx context.Context, id int64) (bool, error) {
	r.ctx = ctx
	r.id = id
	r.assignmentCheckCalls++
	return r.hasActiveAssignments, r.err
}

func (r *employeesRepositoryStub) UpdateEmployee(ctx context.Context, employee domain.Employee) (domain.Employee, error) {
	r.ctx = ctx
	r.employee = employee
	r.updateCalls++
	return r.result, r.err
}

func TestEmployeesServicePassesDataToRepository(t *testing.T) {
	ctx := context.WithValue(context.Background(), struct{}{}, "request")
	repository := &employeesRepositoryStub{
		result: domain.Employee{
			ID:         7,
			FullName:   "Иван Иванов",
			Department: "ИТ",
			IsActive:   true,
		},
		list: []domain.Employee{
			{
				ID:         7,
				FullName:   "Иван Иванов",
				Department: "ИТ",
				IsActive:   true,
			},
		},
	}
	service := NewEmployeesService(repository)

	created, err := service.CreateEmployee(ctx, "Иван Иванов", "ИТ")
	if err != nil || created.ID != 7 || repository.employee.ID != 0 || !repository.employee.IsActive || repository.employee.FullName != "Иван Иванов" || repository.employee.Department != "ИТ" || repository.ctx != ctx || repository.saveCalls != 1 {
		t.Fatalf("CreateEmployee() = %+v, %v; repository input = %+v", created, err, repository.employee)
	}

	listed, err := service.GetEmployees(ctx)
	if err != nil || len(listed) != 1 || listed[0].ID != 7 || repository.listCalls != 1 {
		t.Fatalf("GetEmployees() = %+v, %v", listed, err)
	}

	got, err := service.GetEmployee(ctx, 7)
	if err != nil || got.ID != 7 || repository.id != 7 || repository.getCalls != 1 {
		t.Fatalf("GetEmployee() = %+v, %v; repository id = %d", got, err, repository.id)
	}

	repository.result.IsActive = false
	updated, err := service.UpdateEmployee(ctx, 7, domain.Employee{
		ID:         99,
		FullName:   "Иван Иванов",
		Department: "Отдел разработки",
		IsActive:   false,
	})
	if err != nil || updated.ID != 7 || updated.IsActive || repository.employee.ID != 7 || repository.employee.Department != "Отдел разработки" || repository.assignmentCheckCalls != 1 || repository.updateCalls != 1 || repository.ctx != ctx {
		t.Fatalf("UpdateEmployee() = %+v, %v; repository input = %+v", updated, err, repository.employee)
	}
}

func TestEmployeesServiceRejectsInvalidInputBeforeRepository(t *testing.T) {
	tests := []struct {
		name string
		call func(*EmployeesService) error
	}{
		{
			name: "blank full name",
			call: func(service *EmployeesService) error {
				_, err := service.CreateEmployee(context.Background(), " \t ", "ИТ")
				return err
			},
		},
		{
			name: "full name over 150 characters",
			call: func(service *EmployeesService) error {
				_, err := service.CreateEmployee(context.Background(), strings.Repeat("Я", 151), "ИТ")
				return err
			},
		},
		{
			name: "blank department",
			call: func(service *EmployeesService) error {
				_, err := service.CreateEmployee(context.Background(), "Иван Иванов", " \t ")
				return err
			},
		},
		{
			name: "department over 100 characters",
			call: func(service *EmployeesService) error {
				_, err := service.CreateEmployee(context.Background(), "Иван Иванов", strings.Repeat("Я", 101))
				return err
			},
		},
		{
			name: "get with invalid ID",
			call: func(service *EmployeesService) error {
				_, err := service.GetEmployee(context.Background(), 0)
				return err
			},
		},
		{
			name: "update with invalid ID",
			call: func(service *EmployeesService) error {
				_, err := service.UpdateEmployee(context.Background(), -1, domain.Employee{
					FullName:   "Иван Иванов",
					Department: "ИТ",
				})
				return err
			},
		},
		{
			name: "update with invalid fields",
			call: func(service *EmployeesService) error {
				_, err := service.UpdateEmployee(context.Background(), 7, domain.Employee{
					FullName:   "Иван Иванов",
					Department: " ",
				})
				return err
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &employeesRepositoryStub{}
			err := test.call(NewEmployeesService(repository))
			if !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("error = %v, want invalid argument", err)
			}
			if repository.saveCalls != 0 || repository.listCalls != 0 || repository.getCalls != 0 || repository.assignmentCheckCalls != 0 || repository.updateCalls != 0 {
				t.Fatalf("invalid request reached repository: %+v", repository)
			}
		})
	}
}

func TestEmployeesServiceRejectsDeactivationWithActiveAssignments(t *testing.T) {
	repository := &employeesRepositoryStub{hasActiveAssignments: true}
	service := NewEmployeesService(repository)
	_, err := service.UpdateEmployee(context.Background(), 7, domain.Employee{
		FullName:   "Иван Иванов",
		Department: "ИТ",
		IsActive:   false,
	})
	if !errors.Is(err, core_errors.ErrConflict) || repository.assignmentCheckCalls != 1 || repository.updateCalls != 0 {
		t.Fatalf("error = %v, assignment checks = %d, updates = %d", err, repository.assignmentCheckCalls, repository.updateCalls)
	}
}

func TestEmployeesServiceDoesNotCheckAssignmentsWhenActive(t *testing.T) {
	repository := &employeesRepositoryStub{hasActiveAssignments: true}
	service := NewEmployeesService(repository)
	_, err := service.UpdateEmployee(context.Background(), 7, domain.Employee{
		FullName:   "Иван Иванов",
		Department: "ИТ",
		IsActive:   true,
	})
	if err != nil || repository.assignmentCheckCalls != 0 || repository.updateCalls != 1 || !repository.employee.IsActive {
		t.Fatalf("error = %v, assignment checks = %d, updates = %d", err, repository.assignmentCheckCalls, repository.updateCalls)
	}
}

func TestEmployeesServicePreservesRepositoryErrors(t *testing.T) {
	tests := []struct {
		name string
		call func(*EmployeesService) error
	}{
		{
			name: "create",
			call: func(service *EmployeesService) error {
				_, err := service.CreateEmployee(context.Background(), "Иван Иванов", "ИТ")
				return err
			},
		},
		{
			name: "list",
			call: func(service *EmployeesService) error {
				_, err := service.GetEmployees(context.Background())
				return err
			},
		},
		{
			name: "get",
			call: func(service *EmployeesService) error {
				_, err := service.GetEmployee(context.Background(), 7)
				return err
			},
		},
		{
			name: "assignment check",
			call: func(service *EmployeesService) error {
				_, err := service.UpdateEmployee(context.Background(), 7, domain.Employee{
					FullName:   "Иван Иванов",
					Department: "ИТ",
					IsActive:   false,
				})
				return err
			},
		},
		{
			name: "update",
			call: func(service *EmployeesService) error {
				_, err := service.UpdateEmployee(context.Background(), 7, domain.Employee{
					FullName:   "Иван Иванов",
					Department: "ИТ",
					IsActive:   true,
				})
				return err
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &employeesRepositoryStub{err: core_errors.ErrNotFound}
			err := test.call(NewEmployeesService(repository))
			if !errors.Is(err, core_errors.ErrNotFound) {
				t.Fatalf("error = %v, want not found", err)
			}
		})
	}
}
