package assignment_postgres_repository

import (
	"time"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

type assignmentModel struct {
	ID          int64
	EquipmentID int64
	EmployeeID  int64
	IssuedAt    time.Time
	ReturnedAt  *time.Time
}

func (m *assignmentModel) Scan(row core_postgres_pool.Row) error {
	return row.Scan(
		&m.ID,
		&m.EquipmentID,
		&m.EmployeeID,
		&m.IssuedAt,
		&m.ReturnedAt,
	)
}

func (m assignmentModel) toDomain() domain.Assignment {
	return domain.Assignment{
		ID:          m.ID,
		EquipmentID: m.EquipmentID,
		EmployeeID:  m.EmployeeID,
		IssuedAt:    m.IssuedAt,
		ReturnedAt:  m.ReturnedAt,
	}
}
