package equipment_transport_http

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

type CreateEquipmentRequest struct {
	InventoryNumber string  `json:"inventory_number"`
	SerialNumber    string  `json:"serial_number"`
	CategoryID      int64   `json:"category_id"`
	Manufacturer    string  `json:"manufacturer"`
	Model           string  `json:"model"`
	LocationID      int64   `json:"location_id"`
	PurchaseDate    *string `json:"purchase_date"`
	WarrantyUntil   *string `json:"warranty_until"`
	Notes           *string `json:"notes"`
}

func (request CreateEquipmentRequest) Validate() error {
	if strings.TrimSpace(request.InventoryNumber) == "" || utf8.RuneCountInString(request.InventoryNumber) > 50 {
		return fmt.Errorf("inventory_number must contain 1 to 50 characters")
	}
	if strings.TrimSpace(request.SerialNumber) == "" || utf8.RuneCountInString(request.SerialNumber) > 100 {
		return fmt.Errorf("serial_number must contain 1 to 100 characters")
	}
	return validateEquipmentDetails(request.CategoryID, request.Manufacturer, request.Model, request.LocationID, request.PurchaseDate, request.WarrantyUntil, request.Notes)
}

type UpdateEquipmentRequest struct {
	CategoryID    int64   `json:"category_id"`
	Manufacturer  string  `json:"manufacturer"`
	Model         string  `json:"model"`
	LocationID    int64   `json:"location_id"`
	PurchaseDate  *string `json:"purchase_date"`
	WarrantyUntil *string `json:"warranty_until"`
	Notes         *string `json:"notes"`
}

func (request UpdateEquipmentRequest) Validate() error {
	return validateEquipmentDetails(request.CategoryID, request.Manufacturer, request.Model, request.LocationID, request.PurchaseDate, request.WarrantyUntil, request.Notes)
}

func validateEquipmentDetails(categoryID int64, manufacturer, model string, locationID int64, purchaseDate, warrantyUntil, notes *string) error {
	if categoryID <= 0 {
		return fmt.Errorf("category_id must be positive")
	}
	if strings.TrimSpace(manufacturer) == "" || utf8.RuneCountInString(manufacturer) > 100 {
		return fmt.Errorf("manufacturer must contain 1 to 100 characters")
	}
	if strings.TrimSpace(model) == "" || utf8.RuneCountInString(model) > 100 {
		return fmt.Errorf("model must contain 1 to 100 characters")
	}
	if locationID <= 0 {
		return fmt.Errorf("location_id must be positive")
	}
	if _, err := parseOptionalDate(purchaseDate, "purchase_date"); err != nil {
		return err
	}
	if _, err := parseOptionalDate(warrantyUntil, "warranty_until"); err != nil {
		return err
	}
	if notes != nil && utf8.RuneCountInString(*notes) > 1000 {
		return fmt.Errorf("notes must have at most 1000 characters")
	}
	return nil
}

func parseOptionalDate(value *string, field string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	date, err := time.Parse("2006-01-02", *value)
	if err != nil {
		return nil, fmt.Errorf("%s must be YYYY-MM-DD: %w", field, core_errors.ErrInvalidArgument)
	}
	return &date, nil
}

func equipmentFromCreateRequest(request CreateEquipmentRequest) (domain.Equipment, error) {
	purchaseDate, err := parseOptionalDate(request.PurchaseDate, "purchase_date")
	if err != nil {
		return domain.Equipment{}, err
	}
	warrantyUntil, err := parseOptionalDate(request.WarrantyUntil, "warranty_until")
	if err != nil {
		return domain.Equipment{}, err
	}
	return domain.Equipment{
		InventoryNumber: request.InventoryNumber,
		SerialNumber:    request.SerialNumber,
		CategoryID:      request.CategoryID,
		Manufacturer:    request.Manufacturer,
		Model:           request.Model,
		LocationID:      request.LocationID,
		PurchaseDate:    purchaseDate,
		WarrantyUntil:   warrantyUntil,
		Notes:           request.Notes,
	}, nil
}

func equipmentFromUpdateRequest(request UpdateEquipmentRequest) (domain.Equipment, error) {
	purchaseDate, err := parseOptionalDate(request.PurchaseDate, "purchase_date")
	if err != nil {
		return domain.Equipment{}, err
	}
	warrantyUntil, err := parseOptionalDate(request.WarrantyUntil, "warranty_until")
	if err != nil {
		return domain.Equipment{}, err
	}
	return domain.Equipment{
		CategoryID:    request.CategoryID,
		Manufacturer:  request.Manufacturer,
		Model:         request.Model,
		LocationID:    request.LocationID,
		PurchaseDate:  purchaseDate,
		WarrantyUntil: warrantyUntil,
		Notes:         request.Notes,
	}, nil
}

type EquipmentResponse struct {
	ID              int64                  `json:"id"`
	InventoryNumber string                 `json:"inventory_number"`
	SerialNumber    string                 `json:"serial_number"`
	CategoryID      int64                  `json:"category_id"`
	Manufacturer    string                 `json:"manufacturer"`
	Model           string                 `json:"model"`
	Status          domain.EquipmentStatus `json:"status"`
	LocationID      int64                  `json:"location_id"`
	PurchaseDate    *string                `json:"purchase_date"`
	WarrantyUntil   *string                `json:"warranty_until"`
	Notes           *string                `json:"notes"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

func equipmentDTOFromDomain(equipment domain.Equipment) EquipmentResponse {
	return EquipmentResponse{
		ID:              equipment.ID,
		InventoryNumber: equipment.InventoryNumber,
		SerialNumber:    equipment.SerialNumber,
		CategoryID:      equipment.CategoryID,
		Manufacturer:    equipment.Manufacturer,
		Model:           equipment.Model,
		Status:          equipment.Status,
		LocationID:      equipment.LocationID,
		PurchaseDate:    formatOptionalDate(equipment.PurchaseDate),
		WarrantyUntil:   formatOptionalDate(equipment.WarrantyUntil),
		Notes:           equipment.Notes,
		CreatedAt:       equipment.CreatedAt,
		UpdatedAt:       equipment.UpdatedAt,
	}
}

func formatOptionalDate(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format("2006-01-02")
	return &formatted
}
