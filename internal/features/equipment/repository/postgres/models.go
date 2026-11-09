package equipment_postgres_repository

import (
	"time"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
)

type equipmentModel struct {
	ID              int64
	InventoryNumber string
	SerialNumber    string
	CategoryID      int64
	Manufacturer    string
	Model           string
	Status          string
	LocationID      int64
	PurchaseDate    *time.Time
	WarrantyUntil   *time.Time
	Notes           *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (m *equipmentModel) Scan(row core_postgres_pool.Row) error {
	return row.Scan(
		&m.ID,
		&m.InventoryNumber,
		&m.SerialNumber,
		&m.CategoryID,
		&m.Manufacturer,
		&m.Model,
		&m.Status,
		&m.LocationID,
		&m.PurchaseDate,
		&m.WarrantyUntil,
		&m.Notes,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
}

func (m equipmentModel) toDomain() domain.Equipment {
	return domain.Equipment{
		ID:              m.ID,
		InventoryNumber: m.InventoryNumber,
		SerialNumber:    m.SerialNumber,
		CategoryID:      m.CategoryID,
		Manufacturer:    m.Manufacturer,
		Model:           m.Model,
		Status:          domain.EquipmentStatus(m.Status),
		LocationID:      m.LocationID,
		PurchaseDate:    m.PurchaseDate,
		WarrantyUntil:   m.WarrantyUntil,
		Notes:           m.Notes,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}
