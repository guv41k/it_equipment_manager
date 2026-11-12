package catalog_postgres_repository

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

func (r *LocationsRepository) GetLocations(ctx context.Context) ([]domain.Location, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, name, description
	FROM locations
	ORDER BY id ASC;
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("select locations: %w", err)
	}
	defer rows.Close()

	locations := make([]domain.Location, 0)
	for rows.Next() {
		var model locationModel
		if err := model.Scan(rows); err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}
		locations = append(locations, model.toDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate locations: %w", err)
	}
	return locations, nil
}
