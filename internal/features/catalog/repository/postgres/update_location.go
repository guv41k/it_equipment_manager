package catalog_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

func (r *LocationsRepository) UpdateLocation(ctx context.Context, location domain.Location) (domain.Location, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE locations
	SET name = $1, description = $2
	WHERE id = $3
	RETURNING id, name, description;
	`
	var model locationModel
	if err := model.Scan(r.pool.QueryRow(ctx, query, location.Name, location.Description, location.ID)); err != nil {
		switch {
		case errors.Is(err, core_postgres_pool.ErrNoRows):
			return domain.Location{}, fmt.Errorf("location id=%d: %w", location.ID, core_errors.ErrNotFound)
		case errors.Is(err, core_postgres_pool.ErrUniqueViolation):
			return domain.Location{}, fmt.Errorf("location name already exists: %w", core_errors.ErrConflict)
		default:
			return domain.Location{}, fmt.Errorf("update location: %w", err)
		}
	}
	return model.toDomain(), nil
}
