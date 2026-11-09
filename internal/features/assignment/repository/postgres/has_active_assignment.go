package assignment_postgres_repository

import (
	"context"
	"fmt"
)

func (r *AssignmentsRepository) HasActiveAssignment(ctx context.Context, equipmentID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT EXISTS (
		SELECT 1 FROM assignments
		WHERE equipment_id = $1 AND returned_at IS NULL
	);
	`
	var hasActive bool
	if err := r.pool.QueryRow(ctx, query, equipmentID).Scan(&hasActive); err != nil {
		return false, fmt.Errorf("check active assignment: %w", err)
	}
	return hasActive, nil
}
