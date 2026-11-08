package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

type EquipmentStatus string

const (
	EquipmentStatusInStock  EquipmentStatus = "in_stock"
	EquipmentStatusAssigned EquipmentStatus = "assigned"
	EquipmentStatusRetired  EquipmentStatus = "retired"
)

type Equipment struct {
	ID              int64
	InventoryNumber string
	SerialNumber    string
	CategoryID      int64
	Manufacturer    string
	Model           string
	Status          EquipmentStatus
	LocationID      int64
	PurchaseDate    *time.Time
	WarrantyUntil   *time.Time
	Notes           *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (e Equipment) Validate() error {
	if strings.TrimSpace(e.InventoryNumber) == "" || utf8.RuneCountInString(e.InventoryNumber) > 50 {
		return fmt.Errorf("inventory number must contain 1 to 50 characters: %w", core_errors.ErrInvalidArgument)
	}
	if strings.TrimSpace(e.SerialNumber) == "" || utf8.RuneCountInString(e.SerialNumber) > 100 {
		return fmt.Errorf("serial number must contain 1 to 100 characters: %w", core_errors.ErrInvalidArgument)
	}
	if e.CategoryID <= 0 {
		return fmt.Errorf("category id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	if strings.TrimSpace(e.Manufacturer) == "" || utf8.RuneCountInString(e.Manufacturer) > 100 {
		return fmt.Errorf("manufacturer must contain 1 to 100 characters: %w", core_errors.ErrInvalidArgument)
	}
	if strings.TrimSpace(e.Model) == "" || utf8.RuneCountInString(e.Model) > 100 {
		return fmt.Errorf("model must contain 1 to 100 characters: %w", core_errors.ErrInvalidArgument)
	}
	switch e.Status {
	case EquipmentStatusInStock, EquipmentStatusAssigned, EquipmentStatusRetired:
	default:
		return fmt.Errorf("invalid equipment status: %w", core_errors.ErrInvalidArgument)
	}
	if e.LocationID <= 0 {
		return fmt.Errorf("location id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	if e.Notes != nil && utf8.RuneCountInString(*e.Notes) > 1000 {
		return fmt.Errorf("notes must have at most 1000 characters: %w", core_errors.ErrInvalidArgument)
	}
	return nil
}
