package catalog_service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

type categoriesRepositoryStub struct {
	calls    int
	ctx      context.Context
	id       int64
	category domain.Category
	result   domain.Category
	list     []domain.Category
	err      error
}

func (r *categoriesRepositoryStub) SaveCategory(ctx context.Context, category domain.Category) (domain.Category, error) {
	r.calls++
	r.ctx = ctx
	r.category = category
	return r.result, r.err
}

func (r *categoriesRepositoryStub) GetCategories(ctx context.Context) ([]domain.Category, error) {
	r.calls++
	r.ctx = ctx
	return r.list, r.err
}

func (r *categoriesRepositoryStub) GetCategory(ctx context.Context, id int64) (domain.Category, error) {
	r.calls++
	r.ctx = ctx
	r.id = id
	return r.result, r.err
}

func (r *categoriesRepositoryStub) UpdateCategory(ctx context.Context, category domain.Category) (domain.Category, error) {
	r.calls++
	r.ctx = ctx
	r.category = category
	return r.result, r.err
}

type locationsRepositoryStub struct {
	calls    int
	ctx      context.Context
	id       int64
	location domain.Location
	result   domain.Location
	list     []domain.Location
	err      error
}

func (r *locationsRepositoryStub) SaveLocation(ctx context.Context, location domain.Location) (domain.Location, error) {
	r.calls++
	r.ctx = ctx
	r.location = location
	return r.result, r.err
}

func (r *locationsRepositoryStub) GetLocations(ctx context.Context) ([]domain.Location, error) {
	r.calls++
	r.ctx = ctx
	return r.list, r.err
}

func (r *locationsRepositoryStub) GetLocation(ctx context.Context, id int64) (domain.Location, error) {
	r.calls++
	r.ctx = ctx
	r.id = id
	return r.result, r.err
}

func (r *locationsRepositoryStub) UpdateLocation(ctx context.Context, location domain.Location) (domain.Location, error) {
	r.calls++
	r.ctx = ctx
	r.location = location
	return r.result, r.err
}

func TestCategoriesServicePassesDataToRepository(t *testing.T) {
	ctx := context.WithValue(context.Background(), struct{}{}, "request")
	description := "Переносное устройство"
	repository := &categoriesRepositoryStub{
		result: domain.Category{
			ID:          7,
			Name:        "Ноутбук",
			Description: &description,
		},
		list: []domain.Category{
			{
				ID:          7,
				Name:        "Ноутбук",
				Description: &description,
			},
		},
	}
	service := NewCategoriesService(repository)

	created, err := service.CreateCategory(ctx, "Ноутбук", &description)
	if err != nil || created.ID != 7 || repository.category.ID != 0 || repository.category.Name != "Ноутбук" || repository.ctx != ctx {
		t.Fatalf("CreateCategory() = %+v, %v; repository input = %+v", created, err, repository.category)
	}

	listed, err := service.GetCategories(ctx)
	if err != nil || len(listed) != 1 || listed[0].ID != 7 {
		t.Fatalf("GetCategories() = %+v, %v", listed, err)
	}

	got, err := service.GetCategory(ctx, 7)
	if err != nil || got.ID != 7 || repository.id != 7 {
		t.Fatalf("GetCategory() = %+v, %v; repository id = %d", got, err, repository.id)
	}

	updated, err := service.UpdateCategory(ctx, 7, domain.Category{
		ID:   99,
		Name: "Ноутбук",
	})
	if err != nil || updated.ID != 7 || repository.category.ID != 7 || repository.category.Description != nil {
		t.Fatalf("UpdateCategory() = %+v, %v; repository input = %+v", updated, err, repository.category)
	}
	if repository.calls != 4 {
		t.Fatalf("repository calls = %d, want 4", repository.calls)
	}
}

func TestLocationsServicePassesDataToRepository(t *testing.T) {
	ctx := context.Background()
	repository := &locationsRepositoryStub{
		result: domain.Location{
			ID:   3,
			Name: "Основной склад",
		},
		list: []domain.Location{
			{
				ID:   3,
				Name: "Основной склад",
			},
		},
	}
	service := NewLocationsService(repository)

	created, err := service.CreateLocation(ctx, "Основной склад", nil)
	if err != nil || created.ID != 3 || repository.location.ID != 0 || repository.location.Name != "Основной склад" || repository.ctx != ctx {
		t.Fatalf("CreateLocation() = %+v, %v; repository input = %+v", created, err, repository.location)
	}

	listed, err := service.GetLocations(ctx)
	if err != nil || len(listed) != 1 || listed[0].ID != 3 {
		t.Fatalf("GetLocations() = %+v, %v", listed, err)
	}

	got, err := service.GetLocation(ctx, 3)
	if err != nil || got.ID != 3 || repository.id != 3 {
		t.Fatalf("GetLocation() = %+v, %v; repository id = %d", got, err, repository.id)
	}

	updated, err := service.UpdateLocation(ctx, 3, domain.Location{
		ID:   99,
		Name: "Основной склад",
	})
	if err != nil || updated.ID != 3 || repository.location.ID != 3 || repository.location.Description != nil {
		t.Fatalf("UpdateLocation() = %+v, %v; repository input = %+v", updated, err, repository.location)
	}
	if repository.calls != 4 {
		t.Fatalf("repository calls = %d, want 4", repository.calls)
	}
}

func TestCatalogServicesRejectInvalidInputBeforeRepository(t *testing.T) {
	longDescription := strings.Repeat("я", 501)
	tests := []struct {
		name string
		call func(*CategoriesService, *LocationsService) error
	}{
		{
			name: "blank category name",
			call: func(categories *CategoriesService, _ *LocationsService) error {
				_, err := categories.CreateCategory(context.Background(), " \t ", nil)
				return err
			},
		},
		{
			name: "category name over 100 characters",
			call: func(categories *CategoriesService, _ *LocationsService) error {
				_, err := categories.CreateCategory(context.Background(), strings.Repeat("я", 101), nil)
				return err
			},
		},
		{
			name: "category description over 500 characters",
			call: func(categories *CategoriesService, _ *LocationsService) error {
				_, err := categories.CreateCategory(context.Background(), "Ноутбук", &longDescription)
				return err
			},
		},
		{
			name: "category ID must be positive",
			call: func(categories *CategoriesService, _ *LocationsService) error {
				_, err := categories.GetCategory(context.Background(), 0)
				return err
			},
		},
		{
			name: "category update validates fields",
			call: func(categories *CategoriesService, _ *LocationsService) error {
				_, err := categories.UpdateCategory(context.Background(), 1, domain.Category{Name: " "})
				return err
			},
		},
		{
			name: "blank location name",
			call: func(_ *CategoriesService, locations *LocationsService) error {
				_, err := locations.CreateLocation(context.Background(), " \t ", nil)
				return err
			},
		},
		{
			name: "location name over 100 characters",
			call: func(_ *CategoriesService, locations *LocationsService) error {
				_, err := locations.CreateLocation(context.Background(), strings.Repeat("я", 101), nil)
				return err
			},
		},
		{
			name: "location description over 500 characters",
			call: func(_ *CategoriesService, locations *LocationsService) error {
				_, err := locations.CreateLocation(context.Background(), "Склад", &longDescription)
				return err
			},
		},
		{
			name: "location ID must be positive",
			call: func(_ *CategoriesService, locations *LocationsService) error {
				_, err := locations.GetLocation(context.Background(), -1)
				return err
			},
		},
		{
			name: "location update validates fields",
			call: func(_ *CategoriesService, locations *LocationsService) error {
				_, err := locations.UpdateLocation(context.Background(), 1, domain.Location{Name: " "})
				return err
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			categoriesRepository := &categoriesRepositoryStub{}
			locationsRepository := &locationsRepositoryStub{}
			err := test.call(NewCategoriesService(categoriesRepository), NewLocationsService(locationsRepository))
			if !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("error = %v, want invalid argument", err)
			}
			if categoriesRepository.calls != 0 || locationsRepository.calls != 0 {
				t.Fatalf("invalid request reached repository: categories = %d, locations = %d", categoriesRepository.calls, locationsRepository.calls)
			}
		})
	}
}

func TestCatalogServicesPreserveRepositoryErrors(t *testing.T) {
	categoryRepository := &categoriesRepositoryStub{err: core_errors.ErrConflict}
	categoryService := NewCategoriesService(categoryRepository)
	_, err := categoryService.CreateCategory(context.Background(), "Ноутбук", nil)
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("CreateCategory() error = %v, want conflict", err)
	}

	locationRepository := &locationsRepositoryStub{err: core_errors.ErrNotFound}
	locationService := NewLocationsService(locationRepository)
	_, err = locationService.GetLocation(context.Background(), 42)
	if !errors.Is(err, core_errors.ErrNotFound) || locationRepository.id != 42 {
		t.Fatalf("GetLocation() error = %v, repository id = %d", err, locationRepository.id)
	}
}

func TestCatalogDomainAcceptsBoundaryLengths(t *testing.T) {
	name := strings.Repeat("я", 100)
	description := strings.Repeat("я", 500)
	category := domain.Category{
		Name:        name,
		Description: &description,
	}
	if err := category.Validate(); err != nil {
		t.Fatalf("Category.Validate() error = %v", err)
	}
	location := domain.Location{
		Name:        name,
		Description: &description,
	}
	if err := location.Validate(); err != nil {
		t.Fatalf("Location.Validate() error = %v", err)
	}
}
