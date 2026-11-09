package equipment_postgres_repository

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_pgx_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool/pgx"
	equipment_service "gitlab.com/guv41k/it_equipment_manager/internal/features/equipment/service"
)

var _ equipment_service.EquipmentRepository = (*EquipmentRepository)(nil)

func TestEquipmentRepositoryPostgres(t *testing.T) {
	dsn := os.Getenv("EQUIPMENT_TEST_DSN")
	if dsn == "" {
		t.Skip("set EQUIPMENT_TEST_DSN to test against a migrated PostgreSQL database")
	}
	address, err := url.Parse(dsn)
	if err != nil || address.Scheme != "postgres" || address.User == nil || address.Port() == "" {
		t.Fatal("EQUIPMENT_TEST_DSN must be a postgres URL with user and port")
	}
	password, _ := address.User.Password()
	ctx := context.Background()
	pool, err := core_pgx_pool.NewPool(ctx, core_pgx_pool.Config{
		Host:     address.Hostname(),
		Port:     address.Port(),
		User:     address.User.Username(),
		Password: password,
		Database: strings.TrimPrefix(address.Path, "/"),
		Timeout:  5 * time.Second,
	})
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	t.Cleanup(pool.Close)

	repository := NewEquipmentRepository(pool)
	prefix := fmt.Sprintf("equipment-test-%d", time.Now().UnixNano())
	var equipmentIDs, categoryIDs, locationIDs []int64
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, id := range equipmentIDs {
			if _, err := pool.Exec(cleanupCtx, "DELETE FROM equipment WHERE id = $1", id); err != nil {
				t.Errorf("delete test equipment %d: %v", id, err)
			}
		}
		for _, id := range categoryIDs {
			if _, err := pool.Exec(cleanupCtx, "DELETE FROM equipment_categories WHERE id = $1", id); err != nil {
				t.Errorf("delete test category %d: %v", id, err)
			}
		}
		for _, id := range locationIDs {
			if _, err := pool.Exec(cleanupCtx, "DELETE FROM locations WHERE id = $1", id); err != nil {
				t.Errorf("delete test location %d: %v", id, err)
			}
		}
	})

	for i := 0; i < 2; i++ {
		var categoryID, locationID int64
		if err := pool.QueryRow(ctx, "INSERT INTO equipment_categories (name) VALUES ($1) RETURNING id", fmt.Sprintf("%s-category-%d", prefix, i)).Scan(&categoryID); err != nil {
			t.Fatalf("insert test category: %v", err)
		}
		categoryIDs = append(categoryIDs, categoryID)
		if err := pool.QueryRow(ctx, "INSERT INTO locations (name) VALUES ($1) RETURNING id", fmt.Sprintf("%s-location-%d", prefix, i)).Scan(&locationID); err != nil {
			t.Fatalf("insert test location: %v", err)
		}
		locationIDs = append(locationIDs, locationID)
	}

	purchaseDate := time.Date(2025, time.March, 2, 0, 0, 0, 0, time.UTC)
	notes := "Новый ноутбук"
	first, err := repository.SaveEquipment(ctx, domain.Equipment{
		InventoryNumber: prefix + "-inv-1",
		SerialNumber:    prefix + "-serial-1",
		CategoryID:      categoryIDs[0],
		Manufacturer:    "Lenovo",
		Model:           "T14",
		Status:          domain.EquipmentStatusInStock,
		LocationID:      locationIDs[0],
		PurchaseDate:    &purchaseDate,
		Notes:           &notes,
	})
	if err != nil || first.ID <= 0 || first.Status != domain.EquipmentStatusInStock || first.PurchaseDate == nil || first.PurchaseDate.Format("2006-01-02") != "2025-03-02" || first.CreatedAt.IsZero() || first.UpdatedAt.IsZero() {
		t.Fatalf("SaveEquipment() = %+v, %v", first, err)
	}
	equipmentIDs = append(equipmentIDs, first.ID)
	second, err := repository.SaveEquipment(ctx, domain.Equipment{
		InventoryNumber: prefix + "-inv-2",
		SerialNumber:    prefix + "-serial-2",
		CategoryID:      categoryIDs[1],
		Manufacturer:    "Dell",
		Model:           "XPS",
		Status:          domain.EquipmentStatusInStock,
		LocationID:      locationIDs[1],
	})
	if err != nil || second.ID <= first.ID {
		t.Fatalf("second SaveEquipment() = %+v, %v", second, err)
	}
	equipmentIDs = append(equipmentIDs, second.ID)

	got, err := repository.GetEquipment(ctx, first.ID)
	if err != nil || got.ID != first.ID || got.InventoryNumber != first.InventoryNumber || got.Notes == nil || *got.Notes != notes {
		t.Fatalf("GetEquipment() = %+v, %v", got, err)
	}
	list, err := repository.GetEquipmentList(ctx, nil, nil, nil, nil, nil)
	if err != nil || !containsOrderedEquipment(list, first.ID, second.ID) {
		t.Fatalf("GetEquipmentList() = %+v, %v", list, err)
	}
	secondOffset := 0
	for i, item := range list {
		if item.ID == second.ID {
			secondOffset = i
			break
		}
	}
	status := domain.EquipmentStatusInStock
	list, err = repository.GetEquipmentList(ctx, &categoryIDs[0], &status, &locationIDs[0], nil, nil)
	if err != nil || len(list) != 1 || list[0].ID != first.ID {
		t.Fatalf("filtered GetEquipmentList() = %+v, %v", list, err)
	}
	limit, offset := 1, secondOffset
	list, err = repository.GetEquipmentList(ctx, nil, nil, nil, &limit, &offset)
	if err != nil || len(list) != 1 || list[0].ID != second.ID {
		t.Fatalf("paged GetEquipmentList() = %+v, %v", list, err)
	}
	missingCategory := second.CategoryID + 1000000
	list, err = repository.GetEquipmentList(ctx, &missingCategory, nil, nil, nil, nil)
	if err != nil || len(list) != 0 {
		t.Fatalf("empty GetEquipmentList() = %+v, %v", list, err)
	}

	_, err = repository.SaveEquipment(ctx, domain.Equipment{
		InventoryNumber: first.InventoryNumber,
		SerialNumber:    prefix + "-serial-3",
		CategoryID:      categoryIDs[0],
		Manufacturer:    "Lenovo",
		Model:           "T14",
		Status:          domain.EquipmentStatusInStock,
		LocationID:      locationIDs[0],
	})
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("duplicate inventory error = %v, want conflict", err)
	}
	_, err = repository.SaveEquipment(ctx, domain.Equipment{
		InventoryNumber: prefix + "-inv-3",
		SerialNumber:    first.SerialNumber,
		CategoryID:      categoryIDs[0],
		Manufacturer:    "Lenovo",
		Model:           "T14",
		Status:          domain.EquipmentStatusInStock,
		LocationID:      locationIDs[0],
	})
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("duplicate serial error = %v, want conflict", err)
	}
	_, err = repository.SaveEquipment(ctx, domain.Equipment{
		InventoryNumber: prefix + "-inv-4",
		SerialNumber:    prefix + "-serial-4",
		CategoryID:      -1,
		Manufacturer:    "Lenovo",
		Model:           "T14",
		Status:          domain.EquipmentStatusInStock,
		LocationID:      locationIDs[0],
	})
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("missing category error = %v, want invalid argument", err)
	}

	if _, err := pool.Exec(ctx, "UPDATE equipment SET updated_at = created_at - INTERVAL '1 day' WHERE id = $1", first.ID); err != nil {
		t.Fatalf("set old updated_at: %v", err)
	}
	before, err := repository.GetEquipment(ctx, first.ID)
	if err != nil {
		t.Fatalf("get equipment before update: %v", err)
	}
	warrantyUntil := time.Date(2028, time.March, 2, 0, 0, 0, 0, time.UTC)
	updated, err := repository.UpdateEquipment(ctx, domain.Equipment{
		ID:              first.ID,
		InventoryNumber: "ignored-inventory",
		SerialNumber:    "ignored-serial",
		CategoryID:      categoryIDs[1],
		Manufacturer:    "HP",
		Model:           "EliteBook",
		Status:          domain.EquipmentStatusRetired,
		LocationID:      locationIDs[1],
		WarrantyUntil:   &warrantyUntil,
	})
	if err != nil || updated.InventoryNumber != first.InventoryNumber || updated.SerialNumber != first.SerialNumber || updated.Status != domain.EquipmentStatusInStock || updated.CategoryID != categoryIDs[1] || updated.LocationID != locationIDs[1] || updated.PurchaseDate != nil || updated.WarrantyUntil == nil || updated.WarrantyUntil.Format("2006-01-02") != "2028-03-02" || updated.Notes != nil || !updated.CreatedAt.Equal(first.CreatedAt) || !updated.UpdatedAt.After(before.UpdatedAt) {
		t.Fatalf("UpdateEquipment() = %+v, %v", updated, err)
	}
	got, err = repository.GetEquipment(ctx, first.ID)
	if err != nil || got.Model != "EliteBook" || got.Status != domain.EquipmentStatusInStock {
		t.Fatalf("GetEquipment() after update = %+v, %v", got, err)
	}

	_, err = repository.GetEquipment(ctx, -1)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("missing GetEquipment() error = %v, want not found", err)
	}
	_, err = repository.UpdateEquipment(ctx, domain.Equipment{
		ID:           -1,
		CategoryID:   categoryIDs[0],
		Manufacturer: "Lenovo",
		Model:        "T14",
		LocationID:   locationIDs[0],
	})
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("missing UpdateEquipment() error = %v, want not found", err)
	}
	_, err = repository.UpdateEquipment(ctx, domain.Equipment{
		ID:           first.ID,
		CategoryID:   -1,
		Manufacturer: "Lenovo",
		Model:        "T14",
		LocationID:   locationIDs[0],
	})
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("missing category on update error = %v, want invalid argument", err)
	}
}

func containsOrderedEquipment(equipment []domain.Equipment, firstID, secondID int64) bool {
	firstIndex, secondIndex := -1, -1
	for i, item := range equipment {
		if item.ID == firstID {
			firstIndex = i
		}
		if item.ID == secondID {
			secondIndex = i
		}
	}
	return firstIndex >= 0 && secondIndex > firstIndex
}
