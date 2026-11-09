package equipment_service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

type equipmentRepositoryStub struct {
	ctx         context.Context
	id          int64
	stored      domain.Equipment
	saved       domain.Equipment
	updated     domain.Equipment
	list        []domain.Equipment
	saveCalls   int
	listCalls   int
	getCalls    int
	updateCalls int
	categoryID  *int64
	status      *domain.EquipmentStatus
	locationID  *int64
	limit       *int
	offset      *int
	saveErr     error
	listErr     error
	getErr      error
	updateErr   error
}

func (r *equipmentRepositoryStub) SaveEquipment(ctx context.Context, equipment domain.Equipment) (domain.Equipment, error) {
	r.ctx = ctx
	r.saved = equipment
	r.saveCalls++
	equipment.ID = 11
	return equipment, r.saveErr
}

func (r *equipmentRepositoryStub) GetEquipmentList(ctx context.Context, categoryID *int64, status *domain.EquipmentStatus, locationID *int64, limit, offset *int) ([]domain.Equipment, error) {
	r.ctx = ctx
	r.categoryID = categoryID
	r.status = status
	r.locationID = locationID
	r.limit = limit
	r.offset = offset
	r.listCalls++
	return r.list, r.listErr
}

func (r *equipmentRepositoryStub) GetEquipment(ctx context.Context, id int64) (domain.Equipment, error) {
	r.ctx = ctx
	r.id = id
	r.getCalls++
	return r.stored, r.getErr
}

func (r *equipmentRepositoryStub) UpdateEquipment(ctx context.Context, equipment domain.Equipment) (domain.Equipment, error) {
	r.ctx = ctx
	r.updated = equipment
	r.updateCalls++
	return equipment, r.updateErr
}

func validEquipment() domain.Equipment {
	return domain.Equipment{
		ID:              11,
		InventoryNumber: "INV-11",
		SerialNumber:    "SER-11",
		CategoryID:      2,
		Manufacturer:    "Lenovo",
		Model:           "T14",
		Status:          domain.EquipmentStatusAssigned,
		LocationID:      3,
		CreatedAt:       time.Date(2025, time.March, 3, 10, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2025, time.March, 3, 10, 0, 0, 0, time.UTC),
	}
}

func TestCreateEquipmentSetsServerFields(t *testing.T) {
	ctx := context.WithValue(context.Background(), struct{}{}, "request")
	repository := &equipmentRepositoryStub{}
	input := validEquipment()
	input.Status = domain.EquipmentStatusRetired
	created, err := NewEquipmentService(repository).CreateEquipment(ctx, input)
	if err != nil || created.ID != 11 || created.Status != domain.EquipmentStatusInStock {
		t.Fatalf("CreateEquipment() = %+v, %v", created, err)
	}
	if repository.saveCalls != 1 || repository.ctx != ctx || repository.saved.ID != 0 || repository.saved.Status != domain.EquipmentStatusInStock || !repository.saved.CreatedAt.IsZero() || !repository.saved.UpdatedAt.IsZero() {
		t.Fatalf("repository input = %+v", repository.saved)
	}
}

func TestCreateEquipmentRejectsInvalidFields(t *testing.T) {
	longNotes := strings.Repeat("Я", 1001)
	tests := []struct {
		name   string
		change func(*domain.Equipment)
	}{
		{
			name: "blank inventory number",
			change: func(e *domain.Equipment) {
				e.InventoryNumber = " \t "
			},
		},
		{
			name: "serial number too long",
			change: func(e *domain.Equipment) {
				e.SerialNumber = strings.Repeat("Я", 101)
			},
		},
		{
			name: "invalid category",
			change: func(e *domain.Equipment) {
				e.CategoryID = 0
			},
		},
		{
			name: "blank manufacturer",
			change: func(e *domain.Equipment) {
				e.Manufacturer = " "
			},
		},
		{
			name: "blank model",
			change: func(e *domain.Equipment) {
				e.Model = " "
			},
		},
		{
			name: "invalid location",
			change: func(e *domain.Equipment) {
				e.LocationID = -1
			},
		},
		{
			name: "notes too long",
			change: func(e *domain.Equipment) {
				e.Notes = &longNotes
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &equipmentRepositoryStub{}
			input := validEquipment()
			test.change(&input)
			_, err := NewEquipmentService(repository).CreateEquipment(context.Background(), input)
			if !errors.Is(err, core_errors.ErrInvalidArgument) || repository.saveCalls != 0 {
				t.Fatalf("error = %v, save calls = %d", err, repository.saveCalls)
			}
		})
	}
}

func TestGetEquipmentListPassesFilters(t *testing.T) {
	ctx := context.WithValue(context.Background(), struct{}{}, "request")
	categoryID, locationID := int64(2), int64(3)
	status := domain.EquipmentStatusInStock
	limit, offset := 10, 5
	repository := &equipmentRepositoryStub{
		list: []domain.Equipment{validEquipment()},
	}
	result, err := NewEquipmentService(repository).GetEquipmentList(ctx, &categoryID, &status, &locationID, &limit, &offset)
	if err != nil || len(result) != 1 || repository.listCalls != 1 || repository.ctx != ctx || repository.categoryID != &categoryID || repository.status != &status || repository.locationID != &locationID || repository.limit != &limit || repository.offset != &offset {
		t.Fatalf("GetEquipmentList() = %+v, %v; repository = %+v", result, err, repository)
	}
	result, err = NewEquipmentService(repository).GetEquipmentList(ctx, nil, nil, nil, nil, nil)
	if err != nil || len(result) != 1 || repository.listCalls != 2 {
		t.Fatalf("unfiltered GetEquipmentList() = %+v, %v", result, err)
	}
}

func TestGetEquipmentListRejectsInvalidFilters(t *testing.T) {
	zeroID, negativeID := int64(0), int64(-1)
	invalidStatus := domain.EquipmentStatus("broken")
	zeroLimit, negativeOffset := 0, -1
	tests := []struct {
		name       string
		categoryID *int64
		status     *domain.EquipmentStatus
		locationID *int64
		limit      *int
		offset     *int
	}{
		{
			name:       "category ID",
			categoryID: &zeroID,
		},
		{
			name:   "status",
			status: &invalidStatus,
		},
		{
			name:       "location ID",
			locationID: &negativeID,
		},
		{
			name:  "limit",
			limit: &zeroLimit,
		},
		{
			name:   "offset",
			offset: &negativeOffset,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &equipmentRepositoryStub{}
			_, err := NewEquipmentService(repository).GetEquipmentList(context.Background(), test.categoryID, test.status, test.locationID, test.limit, test.offset)
			if !errors.Is(err, core_errors.ErrInvalidArgument) || repository.listCalls != 0 {
				t.Fatalf("error = %v, list calls = %d", err, repository.listCalls)
			}
		})
	}
}

func TestGetEquipmentRejectsInvalidID(t *testing.T) {
	repository := &equipmentRepositoryStub{}
	_, err := NewEquipmentService(repository).GetEquipment(context.Background(), 0)
	if !errors.Is(err, core_errors.ErrInvalidArgument) || repository.getCalls != 0 {
		t.Fatalf("error = %v, get calls = %d", err, repository.getCalls)
	}
}

func TestUpdateEquipmentPreservesIdentifiersAndStatus(t *testing.T) {
	ctx := context.WithValue(context.Background(), struct{}{}, "request")
	repository := &equipmentRepositoryStub{stored: validEquipment()}
	input := domain.Equipment{
		ID:              999,
		InventoryNumber: "OTHER-INV",
		SerialNumber:    "OTHER-SER",
		CategoryID:      4,
		Manufacturer:    "Dell",
		Model:           "XPS",
		Status:          domain.EquipmentStatusRetired,
		LocationID:      5,
		CreatedAt:       time.Now(),
	}
	updated, err := NewEquipmentService(repository).UpdateEquipment(ctx, 11, input)
	if err != nil || repository.getCalls != 1 || repository.updateCalls != 1 || repository.id != 11 || repository.ctx != ctx {
		t.Fatalf("UpdateEquipment() = %+v, %v; repository = %+v", updated, err, repository)
	}
	if updated.ID != 11 || updated.InventoryNumber != "INV-11" || updated.SerialNumber != "SER-11" || updated.Status != domain.EquipmentStatusAssigned || updated.CategoryID != 4 || updated.Manufacturer != "Dell" || updated.Model != "XPS" || updated.LocationID != 5 || updated.Notes != nil || updated.PurchaseDate != nil || !updated.CreatedAt.Equal(repository.stored.CreatedAt) {
		t.Fatalf("updated equipment = %+v", updated)
	}
}

func TestUpdateEquipmentStopsOnInvalidData(t *testing.T) {
	repository := &equipmentRepositoryStub{stored: validEquipment()}
	input := domain.Equipment{
		CategoryID:   2,
		Manufacturer: " ",
		Model:        "T14",
		LocationID:   3,
	}
	_, err := NewEquipmentService(repository).UpdateEquipment(context.Background(), 11, input)
	if !errors.Is(err, core_errors.ErrInvalidArgument) || repository.getCalls != 1 || repository.updateCalls != 0 {
		t.Fatalf("error = %v, get calls = %d, update calls = %d", err, repository.getCalls, repository.updateCalls)
	}
	_, err = NewEquipmentService(repository).UpdateEquipment(context.Background(), 0, input)
	if !errors.Is(err, core_errors.ErrInvalidArgument) || repository.getCalls != 1 {
		t.Fatalf("invalid ID error = %v, get calls = %d", err, repository.getCalls)
	}
}

func TestEquipmentServicePreservesRepositoryErrors(t *testing.T) {
	tests := []struct {
		name       string
		repository equipmentRepositoryStub
		call       func(*EquipmentService) error
		want       error
	}{
		{
			name: "create conflict",
			repository: equipmentRepositoryStub{
				saveErr: core_errors.ErrConflict,
			},
			call: func(service *EquipmentService) error {
				_, err := service.CreateEquipment(context.Background(), validEquipment())
				return err
			},
			want: core_errors.ErrConflict,
		},
		{
			name: "list error",
			repository: equipmentRepositoryStub{
				listErr: core_errors.ErrNotFound,
			},
			call: func(service *EquipmentService) error {
				_, err := service.GetEquipmentList(context.Background(), nil, nil, nil, nil, nil)
				return err
			},
			want: core_errors.ErrNotFound,
		},
		{
			name: "get missing",
			repository: equipmentRepositoryStub{
				getErr: core_errors.ErrNotFound,
			},
			call: func(service *EquipmentService) error {
				_, err := service.GetEquipment(context.Background(), 11)
				return err
			},
			want: core_errors.ErrNotFound,
		},
		{
			name: "update missing",
			repository: equipmentRepositoryStub{
				getErr: core_errors.ErrNotFound,
			},
			call: func(service *EquipmentService) error {
				_, err := service.UpdateEquipment(context.Background(), 11, validEquipment())
				return err
			},
			want: core_errors.ErrNotFound,
		},
		{
			name: "update error",
			repository: equipmentRepositoryStub{
				stored:    validEquipment(),
				updateErr: core_errors.ErrConflict,
			},
			call: func(service *EquipmentService) error {
				_, err := service.UpdateEquipment(context.Background(), 11, validEquipment())
				return err
			},
			want: core_errors.ErrConflict,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.call(NewEquipmentService(&test.repository))
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}
