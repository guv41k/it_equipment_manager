package employee_postgres_repository

import (
	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

type employeeModel struct {
	ID         int64
	FullName   string
	Department string
	IsActive   bool
}

func (m *employeeModel) Scan(row core_postgres_pool.Row) error {
	return row.Scan(&m.ID, &m.FullName, &m.Department, &m.IsActive)
}

func (m employeeModel) toDomain() domain.Employee {
	return domain.Employee{
		ID:         m.ID,
		FullName:   m.FullName,
		Department: m.Department,
		IsActive:   m.IsActive,
	}
}
