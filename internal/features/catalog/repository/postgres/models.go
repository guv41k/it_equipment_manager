package catalog_postgres_repository

import (
	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

type categoryModel struct {
	ID          int64
	Name        string
	Description *string
}

func (m *categoryModel) Scan(row core_postgres_pool.Row) error {
	return row.Scan(&m.ID, &m.Name, &m.Description)
}

func (m categoryModel) toDomain() domain.Category {
	return domain.Category{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
	}
}

type locationModel struct {
	ID          int64
	Name        string
	Description *string
}

func (m *locationModel) Scan(row core_postgres_pool.Row) error {
	return row.Scan(&m.ID, &m.Name, &m.Description)
}

func (m locationModel) toDomain() domain.Location {
	return domain.Location{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
	}
}
