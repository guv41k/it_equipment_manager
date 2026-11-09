package assignment_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

func (s *AssignmentsService) IssueEquipment(ctx context.Context, equipmentID, employeeID int64) (domain.Assignment, error) {
	assignment := domain.Assignment{
		EquipmentID: equipmentID,
		EmployeeID:  employeeID,
	}
	if err := assignment.Validate(); err != nil {
		return domain.Assignment{}, fmt.Errorf("validate assignment: %w", err)
	}
	equipment, err := s.assignmentsRepository.GetEquipment(ctx, equipmentID)
	if err != nil {
		return domain.Assignment{}, fmt.Errorf("get equipment for issue: %w", err)
	}
	if equipment.Status != domain.EquipmentStatusInStock {
		return domain.Assignment{}, fmt.Errorf("equipment is not in stock: %w", core_errors.ErrConflict)
	}
	employee, err := s.assignmentsRepository.GetEmployee(ctx, employeeID)
	if err != nil {
		return domain.Assignment{}, fmt.Errorf("get employee for issue: %w", err)
	}
	if !employee.IsActive {
		return domain.Assignment{}, fmt.Errorf("employee is inactive: %w", core_errors.ErrConflict)
	}
	hasActiveAssignment, err := s.assignmentsRepository.HasActiveAssignment(ctx, equipmentID)
	if err != nil {
		return domain.Assignment{}, fmt.Errorf("check active assignment: %w", err)
	}
	if hasActiveAssignment {
		return domain.Assignment{}, fmt.Errorf("equipment has active assignment: %w", core_errors.ErrConflict)
	}
	issued, err := s.assignmentsRepository.SaveAssignment(ctx, assignment)
	if err != nil {
		return domain.Assignment{}, fmt.Errorf("save assignment in repository: %w", err)
	}
	return issued, nil
}
