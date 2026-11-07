package catalog_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

func (r *CategoriesRepository) GetCategory(ctx context.Context, id int64) (domain.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, name, description
	FROM equipment_categories
	WHERE id = $1;
	`
	var model categoryModel
	if err := model.Scan(r.pool.QueryRow(ctx, query, id)); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Category{}, fmt.Errorf("category id=%d: %w", id, core_errors.ErrNotFound)
		}
		return domain.Category{}, fmt.Errorf("select category: %w", err)
	}
	return model.toDomain(), nil
}
