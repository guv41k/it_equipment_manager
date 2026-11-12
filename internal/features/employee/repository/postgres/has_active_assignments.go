package employee_postgres_repository

import (
	"context"
	"fmt"
)

func (r *EmployeesRepository) HasActiveAssignments(ctx context.Context, id int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT EXISTS (
		SELECT 1
		FROM assignments
		WHERE employee_id = $1 AND returned_at IS NULL
	);
	`
	var hasAssignments bool
	if err := r.pool.QueryRow(ctx, query, id).Scan(&hasAssignments); err != nil {
		return false, fmt.Errorf("check active assignments: %w", err)
	}
	return hasAssignments, nil
}
