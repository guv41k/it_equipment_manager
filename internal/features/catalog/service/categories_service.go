package catalog_service

import (
	"context"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

type CategoriesRepository interface {
	SaveCategory(context.Context, domain.Category) (domain.Category, error)
	GetCategories(context.Context) ([]domain.Category, error)
	GetCategory(context.Context, int64) (domain.Category, error)
	UpdateCategory(context.Context, domain.Category) (domain.Category, error)
}

type CategoriesService struct {
	categoriesRepository CategoriesRepository
}

func NewCategoriesService(repository CategoriesRepository) *CategoriesService {
	return &CategoriesService{
		categoriesRepository: repository,
	}
}
