package catalog_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

func (r *LocationsRepository) GetLocation(ctx context.Context, id int64) (domain.Location, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, name, description
	FROM locations
	WHERE id = $1;
	`
	var model locationModel
	if err := model.Scan(r.pool.QueryRow(ctx, query, id)); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Location{}, fmt.Errorf("location id=%d: %w", id, core_errors.ErrNotFound)
		}
		return domain.Location{}, fmt.Errorf("select location: %w", err)
	}
	return model.toDomain(), nil
}
