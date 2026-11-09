package assignment_postgres_repository

import (
	"context"
	"fmt"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

func (r *AssignmentsRepository) GetEquipmentAssignments(ctx context.Context, equipmentID int64) ([]domain.Assignment, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, equipment_id, employee_id, issued_at, returned_at
	FROM assignments
	WHERE equipment_id = $1
	ORDER BY issued_at ASC, id ASC;
	`
	rows, err := r.pool.Query(ctx, query, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("select equipment assignments: %w", err)
	}
	defer rows.Close()

	assignments := make([]domain.Assignment, 0)
	for rows.Next() {
		var model assignmentModel
		if err := model.Scan(rows); err != nil {
			return nil, fmt.Errorf("scan equipment assignment: %w", err)
		}
		assignments = append(assignments, model.toDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate equipment assignments: %w", err)
	}
	return assignments, nil
}
