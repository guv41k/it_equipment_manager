package catalog_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

func (s *LocationsService) UpdateLocation(ctx context.Context, id int64, location domain.Location) (domain.Location, error) {
	if id <= 0 {
		return domain.Location{}, fmt.Errorf("location id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	location.ID = id
	if err := location.Validate(); err != nil {
		return domain.Location{}, fmt.Errorf("validate location: %w", err)
	}
	updated, err := s.locationsRepository.UpdateLocation(ctx, location)
	if err != nil {
		return domain.Location{}, fmt.Errorf("update location in repository: %w", err)
	}
	return updated, nil
}
