package catalog_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

func (s *LocationsService) GetLocations(ctx context.Context) ([]domain.Location, error) {
	locations, err := s.locationsRepository.GetLocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("get locations from repository: %w", err)
	}
	return locations, nil
}
