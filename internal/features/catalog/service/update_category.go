package catalog_service

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

func (s *CategoriesService) UpdateCategory(ctx context.Context, id int64, category domain.Category) (domain.Category, error) {
	if id <= 0 {
		return domain.Category{}, fmt.Errorf("category id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	category.ID = id
	if err := category.Validate(); err != nil {
		return domain.Category{}, fmt.Errorf("validate category: %w", err)
	}
	updated, err := s.categoriesRepository.UpdateCategory(ctx, category)
	if err != nil {
		return domain.Category{}, fmt.Errorf("update category in repository: %w", err)
	}
	return updated, nil
}
