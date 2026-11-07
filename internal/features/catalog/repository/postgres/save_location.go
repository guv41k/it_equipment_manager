package catalog_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

func (r *LocationsRepository) SaveLocation(ctx context.Context, location domain.Location) (domain.Location, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	INSERT INTO locations (name, description)
	VALUES ($1, $2)
	RETURNING id, name, description;
	`
	var model locationModel
	if err := model.Scan(r.pool.QueryRow(ctx, query, location.Name, location.Description)); err != nil {
		if errors.Is(err, core_postgres_pool.ErrUniqueViolation) {
			return domain.Location{}, fmt.Errorf("location name already exists: %w", core_errors.ErrConflict)
		}
		return domain.Location{}, fmt.Errorf("insert location: %w", err)
	}
	return model.toDomain(), nil
}
