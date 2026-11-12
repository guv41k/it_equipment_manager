package catalog_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

func (s *CategoriesService) GetCategories(ctx context.Context) ([]domain.Category, error) {
	categories, err := s.categoriesRepository.GetCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("get categories from repository: %w", err)
	}
	return categories, nil
}
