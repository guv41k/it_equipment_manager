package catalog_postgres_repository

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
	catalog_service "gitlab.com/guv41k/it_equipment_manager/internal/features/catalog/service"
)

var (
	_ catalog_service.CategoriesRepository = (*CategoriesRepository)(nil)
	_ catalog_service.LocationsRepository  = (*LocationsRepository)(nil)
)
func TestCatalogRepositoryPostgres(t *testing.T) {
	dsn := os.Getenv("CATALOG_TEST_DSN")
	if dsn == "" {
		t.Skip("set CATALOG_TEST_DSN to test against a migrated PostgreSQL database")
	}
	address, err := url.Parse(dsn)
	if err != nil || address.Scheme != "postgres" || address.User == nil || address.Port() == "" {
		t.Fatal("CATALOG_TEST_DSN must be a postgres URL with user and port")
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

	categories := NewCategoriesRepository(pool)
	locations := NewLocationsRepository(pool)
	prefix := fmt.Sprintf("catalog-test-%d", time.Now().UnixNano())
	var categoryIDs, locationIDs []int64
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
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

	category, err := categories.SaveCategory(ctx, domain.Category{Name: prefix + "-category-a"})
	if err != nil || category.ID <= 0 || category.Description != nil {
		t.Fatalf("SaveCategory() = %+v, %v", category, err)
	}
	categoryIDs = append(categoryIDs, category.ID)
	description := "Описание"
	secondCategory, err := categories.SaveCategory(ctx, domain.Category{
		Name:        prefix + "-category-b",
		Description: &description,
	})
	if err != nil || secondCategory.ID <= category.ID {
		t.Fatalf("second SaveCategory() = %+v, %v", secondCategory, err)
	}
	categoryIDs = append(categoryIDs, secondCategory.ID)
	gotCategory, err := categories.GetCategory(ctx, category.ID)
	if err != nil || gotCategory.Name != category.Name || gotCategory.Description != nil {
		t.Fatalf("GetCategory() = %+v, %v", gotCategory, err)
	}
	categoryList, err := categories.GetCategories(ctx)
	if err != nil || !containsOrderedCategories(categoryList, category.ID, secondCategory.ID) {
		t.Fatalf("GetCategories() = %+v, %v", categoryList, err)
	}
	category.Name = prefix + "-category-updated"
	category.Description = &description
	updatedCategory, err := categories.UpdateCategory(ctx, category)
	if err != nil || updatedCategory.Name != category.Name || updatedCategory.Description == nil || *updatedCategory.Description != description {
		t.Fatalf("UpdateCategory() = %+v, %v", updatedCategory, err)
	}
	_, err = categories.SaveCategory(ctx, domain.Category{Name: secondCategory.Name})
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("duplicate SaveCategory() error = %v, want conflict", err)
	}
	category.Name = secondCategory.Name
	_, err = categories.UpdateCategory(ctx, category)
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("duplicate UpdateCategory() error = %v, want conflict", err)
	}
	_, err = categories.GetCategory(ctx, -1)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("missing GetCategory() error = %v, want not found", err)
	}
	_, err = categories.UpdateCategory(ctx, domain.Category{
		ID:   -1,
		Name: prefix + "-missing-category",
	})
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("missing UpdateCategory() error = %v, want not found", err)
	}

	location, err := locations.SaveLocation(ctx, domain.Location{Name: prefix + "-location-a"})
	if err != nil || location.ID <= 0 || location.Description != nil {
		t.Fatalf("SaveLocation() = %+v, %v", location, err)
	}
	locationIDs = append(locationIDs, location.ID)
	secondLocation, err := locations.SaveLocation(ctx, domain.Location{
		Name:        prefix + "-location-b",
		Description: &description,
	})
	if err != nil || secondLocation.ID <= location.ID {
		t.Fatalf("second SaveLocation() = %+v, %v", secondLocation, err)
	}
	locationIDs = append(locationIDs, secondLocation.ID)
	gotLocation, err := locations.GetLocation(ctx, location.ID)
	if err != nil || gotLocation.Name != location.Name || gotLocation.Description != nil {
		t.Fatalf("GetLocation() = %+v, %v", gotLocation, err)
	}
	locationList, err := locations.GetLocations(ctx)
	if err != nil || !containsOrderedLocations(locationList, location.ID, secondLocation.ID) {
		t.Fatalf("GetLocations() = %+v, %v", locationList, err)
	}
	location.Name = prefix + "-location-updated"
	location.Description = &description
	updatedLocation, err := locations.UpdateLocation(ctx, location)
	if err != nil || updatedLocation.Name != location.Name || updatedLocation.Description == nil || *updatedLocation.Description != description {
		t.Fatalf("UpdateLocation() = %+v, %v", updatedLocation, err)
	}
	_, err = locations.SaveLocation(ctx, domain.Location{Name: secondLocation.Name})
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("duplicate SaveLocation() error = %v, want conflict", err)
	}
	location.Name = secondLocation.Name
	_, err = locations.UpdateLocation(ctx, location)
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("duplicate UpdateLocation() error = %v, want conflict", err)
	}
	_, err = locations.GetLocation(ctx, -1)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("missing GetLocation() error = %v, want not found", err)
	}
	_, err = locations.UpdateLocation(ctx, domain.Location{
		ID:   -1,
		Name: prefix + "-missing-location",
	})
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("missing UpdateLocation() error = %v, want not found", err)
	}
}

func containsOrderedCategories(categories []domain.Category, firstID, secondID int64) bool {
	firstIndex, secondIndex := -1, -1
	for i, category := range categories {
		if category.ID == firstID {
			firstIndex = i
		}
		if category.ID == secondID {
			secondIndex = i
		}
	}
	return firstIndex >= 0 && secondIndex > firstIndex
}

func containsOrderedLocations(locations []domain.Location, firstID, secondID int64) bool {
	firstIndex, secondIndex := -1, -1
	for i, location := range locations {
		if location.ID == firstID {
			firstIndex = i
		}
		if location.ID == secondID {
			secondIndex = i
		}
	}
	return firstIndex >= 0 && secondIndex > firstIndex
}
