package catalog_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

func (s *LocationsService) GetLocation(ctx context.Context, id int64) (domain.Location, error) {
	if id <= 0 {
		return domain.Location{}, fmt.Errorf("location id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	location, err := s.locationsRepository.GetLocation(ctx, id)
	if err != nil {
		return domain.Location{}, fmt.Errorf("get location from repository: %w", err)
	}
	return location, nil
}
