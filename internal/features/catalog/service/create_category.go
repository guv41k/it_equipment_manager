package catalog_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

func (s *CategoriesService) CreateCategory(ctx context.Context, name string, description *string) (domain.Category, error) {
	category := domain.Category{
		Name:        name,
		Description: description,
	}
	if err := category.Validate(); err != nil {
		return domain.Category{}, fmt.Errorf("validate category: %w", err)
	}
	category, err := s.categoriesRepository.SaveCategory(ctx, category)
	if err != nil {
		return domain.Category{}, fmt.Errorf("save category in repository: %w", err)
	}
	return category, nil
}
