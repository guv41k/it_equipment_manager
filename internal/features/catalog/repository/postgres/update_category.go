package catalog_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

func (r *CategoriesRepository) UpdateCategory(ctx context.Context, category domain.Category) (domain.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE equipment_categories
	SET name = $1, description = $2
	WHERE id = $3
	RETURNING id, name, description;
	`
	var model categoryModel
	if err := model.Scan(r.pool.QueryRow(ctx, query, category.Name, category.Description, category.ID)); err != nil {
		switch {
		case errors.Is(err, core_postgres_pool.ErrNoRows):
			return domain.Category{}, fmt.Errorf("category id=%d: %w", category.ID, core_errors.ErrNotFound)
		case errors.Is(err, core_postgres_pool.ErrUniqueViolation):
			return domain.Category{}, fmt.Errorf("category name already exists: %w", core_errors.ErrConflict)
		default:
			return domain.Category{}, fmt.Errorf("update category: %w", err)
		}
	}
	return model.toDomain(), nil
}
