package assignment_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

func (s *AssignmentsService) GetEquipmentAssignments(ctx context.Context, equipmentID int64) ([]domain.Assignment, error) {
	if equipmentID <= 0 {
		return nil, fmt.Errorf("equipment id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	if _, err := s.assignmentsRepository.GetEquipment(ctx, equipmentID); err != nil {
		return nil, fmt.Errorf("get equipment for assignments: %w", err)
	}
	assignments, err := s.assignmentsRepository.GetEquipmentAssignments(ctx, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("get equipment assignments from repository: %w", err)
	}
	return assignments, nil
}
