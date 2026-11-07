package catalog_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

func (r *CategoriesRepository) SaveCategory(ctx context.Context, category domain.Category) (domain.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	INSERT INTO equipment_categories (name, description)
	VALUES ($1, $2)
	RETURNING id, name, description;
	`
	var model categoryModel
	if err := model.Scan(r.pool.QueryRow(ctx, query, category.Name, category.Description)); err != nil {
		if errors.Is(err, core_postgres_pool.ErrUniqueViolation) {
			return domain.Category{}, fmt.Errorf("category name already exists: %w", core_errors.ErrConflict)
		}
		return domain.Category{}, fmt.Errorf("insert category: %w", err)
	}
	return model.toDomain(), nil
}
