package assignment_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

type assignmentsRepositoryStub struct {
	ctx               context.Context
	equipmentID       int64
	employeeID        int64
	saved             domain.Assignment
	equipment         domain.Equipment
	employee          domain.Employee
	assignment        domain.Assignment
	history           []domain.Assignment
	hasActive         bool
	getEquipmentCalls int
	getEmployeeCalls  int
	checkActiveCalls  int
	saveCalls         int
	returnCalls       int
	historyCalls      int
	getEquipmentErr   error
	getEmployeeErr    error
	checkActiveErr    error
	saveErr           error
	returnErr         error
	historyErr        error
}

func (r *assignmentsRepositoryStub) GetEquipment(ctx context.Context, id int64) (domain.Equipment, error) {
	r.ctx = ctx
	r.equipmentID = id
	r.getEquipmentCalls++
	return r.equipment, r.getEquipmentErr
}

func (r *assignmentsRepositoryStub) GetEmployee(ctx context.Context, id int64) (domain.Employee, error) {
	r.ctx = ctx
	r.employeeID = id
	r.getEmployeeCalls++
	return r.employee, r.getEmployeeErr
}

func (r *assignmentsRepositoryStub) HasActiveAssignment(ctx context.Context, id int64) (bool, error) {
	r.ctx = ctx
	r.equipmentID = id
	r.checkActiveCalls++
	return r.hasActive, r.checkActiveErr
}

func (r *assignmentsRepositoryStub) SaveAssignment(ctx context.Context, assignment domain.Assignment) (domain.Assignment, error) {
	r.ctx = ctx
	r.saved = assignment
	r.saveCalls++
	return r.assignment, r.saveErr
}

func (r *assignmentsRepositoryStub) ReturnAssignment(ctx context.Context, id int64) (domain.Assignment, error) {
	r.ctx = ctx
	r.equipmentID = id
	r.returnCalls++
	return r.assignment, r.returnErr
}

func (r *assignmentsRepositoryStub) GetEquipmentAssignments(ctx context.Context, id int64) ([]domain.Assignment, error) {
	r.ctx = ctx
	r.equipmentID = id
	r.historyCalls++
	return r.history, r.historyErr
}

func readyAssignmentsRepository() *assignmentsRepositoryStub {
	issuedAt := time.Date(2026, time.November, 9, 10, 0, 0, 0, time.UTC)
	return &assignmentsRepositoryStub{
		equipment: domain.Equipment{
			ID:     11,
			Status: domain.EquipmentStatusInStock,
		},
		employee: domain.Employee{
			ID:       7,
			IsActive: true,
		},
		assignment: domain.Assignment{
			ID:          3,
			EquipmentID: 11,
			EmployeeID:  7,
			IssuedAt:    issuedAt,
		},
	}
}

func TestIssueEquipment(t *testing.T) {
	ctx := context.WithValue(context.Background(), struct{}{}, "request")
	repository := readyAssignmentsRepository()
	issued, err := NewAssignmentsService(repository).IssueEquipment(ctx, 11, 7)
	if err != nil || issued.ID != 3 || repository.saved.EquipmentID != 11 || repository.saved.EmployeeID != 7 || repository.saved.ID != 0 || !repository.saved.IssuedAt.IsZero() || repository.saved.ReturnedAt != nil || repository.ctx != ctx || repository.equipmentID != 11 || repository.employeeID != 7 || repository.getEquipmentCalls != 1 || repository.getEmployeeCalls != 1 || repository.checkActiveCalls != 1 || repository.saveCalls != 1 {
		t.Fatalf("IssueEquipment() = %+v, %v; repository = %+v", issued, err, repository)
	}
}

func TestIssueEquipmentRejectsInvalidIDs(t *testing.T) {
	for _, ids := range [][2]int64{{0, 7}, {11, 0}} {
		repository := readyAssignmentsRepository()
		_, err := NewAssignmentsService(repository).IssueEquipment(context.Background(), ids[0], ids[1])
		if !errors.Is(err, core_errors.ErrInvalidArgument) || repository.getEquipmentCalls != 0 || repository.saveCalls != 0 {
			t.Fatalf("IssueEquipment(%d, %d) error = %v; repository = %+v", ids[0], ids[1], err, repository)
		}
	}
}

func TestIssueEquipmentRejectsUnavailableEquipmentOrEmployee(t *testing.T) {
	tests := []struct {
		name   string
		change func(*assignmentsRepositoryStub)
	}{
		{
			name: "assigned equipment",
			change: func(r *assignmentsRepositoryStub) {
				r.equipment.Status = domain.EquipmentStatusAssigned
			},
		},
		{
			name: "retired equipment",
			change: func(r *assignmentsRepositoryStub) {
				r.equipment.Status = domain.EquipmentStatusRetired
			},
		},
		{
			name: "inactive employee",
			change: func(r *assignmentsRepositoryStub) {
				r.employee.IsActive = false
			},
		},
		{
			name: "active assignment",
			change: func(r *assignmentsRepositoryStub) {
				r.hasActive = true
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := readyAssignmentsRepository()
			test.change(repository)
			_, err := NewAssignmentsService(repository).IssueEquipment(context.Background(), 11, 7)
			if !errors.Is(err, core_errors.ErrConflict) || repository.saveCalls != 0 {
				t.Fatalf("IssueEquipment() error = %v; save calls = %d", err, repository.saveCalls)
			}
		})
	}
}

func TestIssueEquipmentPropagatesRepositoryErrors(t *testing.T) {
	problem := errors.New("repository failure")
	tests := []struct {
		name   string
		change func(*assignmentsRepositoryStub)
	}{
		{
			name: "equipment lookup",
			change: func(r *assignmentsRepositoryStub) {
				r.getEquipmentErr = problem
			},
		},
		{
			name: "employee lookup",
			change: func(r *assignmentsRepositoryStub) {
				r.getEmployeeErr = problem
			},
		},
		{
			name: "active assignment check",
			change: func(r *assignmentsRepositoryStub) {
				r.checkActiveErr = problem
			},
		},
		{
			name: "save",
			change: func(r *assignmentsRepositoryStub) {
				r.saveErr = problem
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := readyAssignmentsRepository()
			test.change(repository)
			_, err := NewAssignmentsService(repository).IssueEquipment(context.Background(), 11, 7)
			if !errors.Is(err, problem) {
				t.Fatalf("IssueEquipment() error = %v", err)
			}
		})
	}
}

func TestReturnEquipment(t *testing.T) {
	repository := readyAssignmentsRepository()
	repository.equipment.Status = domain.EquipmentStatusAssigned
	returnedAt := repository.assignment.IssuedAt.Add(time.Hour)
	repository.assignment.ReturnedAt = &returnedAt
	returned, err := NewAssignmentsService(repository).ReturnEquipment(context.Background(), 11)
	if err != nil || returned.ReturnedAt == nil || !returned.ReturnedAt.Equal(returnedAt) || repository.equipmentID != 11 || repository.getEquipmentCalls != 1 || repository.returnCalls != 1 {
		t.Fatalf("ReturnEquipment() = %+v, %v; repository = %+v", returned, err, repository)
	}
}

func TestReturnEquipmentRejectsInvalidOrUnavailableEquipment(t *testing.T) {
	tests := []struct {
		name   string
		id     int64
		status domain.EquipmentStatus
		want   error
	}{
		{
			name: "invalid id",
			id:   0,
			want: core_errors.ErrInvalidArgument,
		},
		{
			name:   "in stock",
			id:     11,
			status: domain.EquipmentStatusInStock,
			want:   core_errors.ErrConflict,
		},
		{
			name:   "retired",
			id:     11,
			status: domain.EquipmentStatusRetired,
			want:   core_errors.ErrConflict,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := readyAssignmentsRepository()
			repository.equipment.Status = test.status
			_, err := NewAssignmentsService(repository).ReturnEquipment(context.Background(), test.id)
			if !errors.Is(err, test.want) || repository.returnCalls != 0 {
				t.Fatalf("ReturnEquipment() error = %v; return calls = %d", err, repository.returnCalls)
			}
		})
	}
}

func TestReturnEquipmentPropagatesRepositoryErrors(t *testing.T) {
	problem := errors.New("repository failure")
	for _, lookup := range []bool{true, false} {
		repository := readyAssignmentsRepository()
		repository.equipment.Status = domain.EquipmentStatusAssigned
		if lookup {
			repository.getEquipmentErr = problem
		} else {
			repository.returnErr = problem
		}
		_, err := NewAssignmentsService(repository).ReturnEquipment(context.Background(), 11)
		if !errors.Is(err, problem) {
			t.Fatalf("ReturnEquipment() error = %v", err)
		}
	}
}

func TestGetEquipmentAssignments(t *testing.T) {
	repository := readyAssignmentsRepository()
	repository.history = []domain.Assignment{repository.assignment}
	assignments, err := NewAssignmentsService(repository).GetEquipmentAssignments(context.Background(), 11)
	if err != nil || len(assignments) != 1 || assignments[0].ID != 3 || repository.equipmentID != 11 || repository.getEquipmentCalls != 1 || repository.historyCalls != 1 {
		t.Fatalf("GetEquipmentAssignments() = %+v, %v; repository = %+v", assignments, err, repository)
	}
	_, err = NewAssignmentsService(repository).GetEquipmentAssignments(context.Background(), 0)
	if !errors.Is(err, core_errors.ErrInvalidArgument) || repository.historyCalls != 1 {
		t.Fatalf("GetEquipmentAssignments(0) error = %v; history calls = %d", err, repository.historyCalls)
	}
}

func TestGetEquipmentAssignmentsRejectsMissingEquipment(t *testing.T) {
	repository := readyAssignmentsRepository()
	repository.getEquipmentErr = core_errors.ErrNotFound
	_, err := NewAssignmentsService(repository).GetEquipmentAssignments(context.Background(), 11)
	if !errors.Is(err, core_errors.ErrNotFound) || repository.historyCalls != 0 {
		t.Fatalf("GetEquipmentAssignments() error = %v; history calls = %d", err, repository.historyCalls)
	}
}
