package assignment_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

func (s *AssignmentsService) ReturnEquipment(ctx context.Context, equipmentID int64) (domain.Assignment, error) {
	if equipmentID <= 0 {
		return domain.Assignment{}, fmt.Errorf("equipment id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	equipment, err := s.assignmentsRepository.GetEquipment(ctx, equipmentID)
	if err != nil {
		return domain.Assignment{}, fmt.Errorf("get equipment for return: %w", err)
	}
	if equipment.Status != domain.EquipmentStatusAssigned {
		return domain.Assignment{}, fmt.Errorf("equipment is not assigned: %w", core_errors.ErrConflict)
	}
	returned, err := s.assignmentsRepository.ReturnAssignment(ctx, equipmentID)
	if err != nil {
		return domain.Assignment{}, fmt.Errorf("return assignment in repository: %w", err)
	}
	return returned, nil
}
