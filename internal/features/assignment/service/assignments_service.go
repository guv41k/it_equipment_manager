package assignment_service

import (
	"context"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

type AssignmentsRepository interface {
	GetEquipment(context.Context, int64) (domain.Equipment, error)
	GetEmployee(context.Context, int64) (domain.Employee, error)
	HasActiveAssignment(context.Context, int64) (bool, error)
	SaveAssignment(context.Context, domain.Assignment) (domain.Assignment, error)
	ReturnAssignment(context.Context, int64) (domain.Assignment, error)
	GetEquipmentAssignments(context.Context, int64) ([]domain.Assignment, error)
}

type AssignmentsService struct {
	assignmentsRepository AssignmentsRepository
}

func NewAssignmentsService(repository AssignmentsRepository) *AssignmentsService {
	return &AssignmentsService{
		assignmentsRepository: repository,
	}
}
