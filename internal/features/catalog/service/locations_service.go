package catalog_service

import (
	"context"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

type LocationsRepository interface {
	SaveLocation(context.Context, domain.Location) (domain.Location, error)
	GetLocations(context.Context) ([]domain.Location, error)
	GetLocation(context.Context, int64) (domain.Location, error)
	UpdateLocation(context.Context, domain.Location) (domain.Location, error)
}

type LocationsService struct {
	locationsRepository LocationsRepository
}

func NewLocationsService(repository LocationsRepository) *LocationsService {
	return &LocationsService{
		locationsRepository: repository,
	}
}
