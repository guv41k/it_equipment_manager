package catalog_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

func (s *CategoriesService) GetCategory(ctx context.Context, id int64) (domain.Category, error) {
	if id <= 0 {
		return domain.Category{}, fmt.Errorf("category id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	category, err := s.categoriesRepository.GetCategory(ctx, id)
	if err != nil {
		return domain.Category{}, fmt.Errorf("get category from repository: %w", err)
	}
	return category, nil
}
