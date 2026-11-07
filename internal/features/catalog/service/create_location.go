package catalog_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

func (s *LocationsService) CreateLocation(ctx context.Context, name string, description *string) (domain.Location, error) {
	location := domain.Location{
		Name:        name,
		Description: description,
	}
	if err := location.Validate(); err != nil {
		return domain.Location{}, fmt.Errorf("validate location: %w", err)
	}
	location, err := s.locationsRepository.SaveLocation(ctx, location)
	if err != nil {
		return domain.Location{}, fmt.Errorf("save location in repository: %w", err)
	}
	return location, nil
}
