package equipment_postgres_repository_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_pgx_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool/pgx"
	assignment_postgres_repository "gitlab.com/guv41k/it_equipment_manager/internal/features/assignment/repository/postgres"
	equipment_postgres_repository "gitlab.com/guv41k/it_equipment_manager/internal/features/equipment/repository/postgres"
)

type retirementFixture struct {
	pool         *core_pgx_pool.Pool
	prefix       string
	categoryID   int64
	locationID   int64
	employeeID   int64
	equipmentIDs []int64
}

func newRetirementFixture(t *testing.T) *retirementFixture {
	t.Helper()
	dsn := os.Getenv("EQUIPMENT_TEST_DSN")
	if dsn == "" {
		t.Skip("set EQUIPMENT_TEST_DSN to test against a migrated PostgreSQL database")
	}
	address, err := url.Parse(dsn)
	if err != nil || address.Scheme != "postgres" || address.User == nil || address.Port() == "" {
		t.Fatal("EQUIPMENT_TEST_DSN must be a postgres URL with user and port")
	}
	password, _ := address.User.Password()
	pool, err := core_pgx_pool.NewPool(context.Background(), core_pgx_pool.Config{
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
	fixture := &retirementFixture{
		pool:   pool,
		prefix: fmt.Sprintf("retirement-test-%d", time.Now().UnixNano()),
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, id := range fixture.equipmentIDs {
			if _, err := pool.Exec(ctx, "DELETE FROM assignments WHERE equipment_id = $1", id); err != nil {
				t.Errorf("delete test assignments for equipment %d: %v", id, err)
			}
			if _, err := pool.Exec(ctx, "DELETE FROM equipment WHERE id = $1", id); err != nil {
				t.Errorf("delete test equipment %d: %v", id, err)
			}
		}
		if fixture.employeeID != 0 {
			if _, err := pool.Exec(ctx, "DELETE FROM employees WHERE id = $1", fixture.employeeID); err != nil {
				t.Errorf("delete test employee: %v", err)
			}
		}
		if fixture.categoryID != 0 {
			if _, err := pool.Exec(ctx, "DELETE FROM equipment_categories WHERE id = $1", fixture.categoryID); err != nil {
				t.Errorf("delete test category: %v", err)
			}
		}
		if fixture.locationID != 0 {
			if _, err := pool.Exec(ctx, "DELETE FROM locations WHERE id = $1", fixture.locationID); err != nil {
				t.Errorf("delete test location: %v", err)
			}
		}
	})
	ctx := context.Background()
	if err := pool.QueryRow(ctx, "INSERT INTO equipment_categories (name) VALUES ($1) RETURNING id", fixture.prefix).Scan(&fixture.categoryID); err != nil {
		t.Fatalf("insert test category: %v", err)
	}
	if err := pool.QueryRow(ctx, "INSERT INTO locations (name) VALUES ($1) RETURNING id", fixture.prefix).Scan(&fixture.locationID); err != nil {
		t.Fatalf("insert test location: %v", err)
	}
	if err := pool.QueryRow(ctx, "INSERT INTO employees (full_name, department) VALUES ($1, 'ИТ') RETURNING id", fixture.prefix).Scan(&fixture.employeeID); err != nil {
		t.Fatalf("insert test employee: %v", err)
	}
	return fixture
}

func (f *retirementFixture) addEquipment(t *testing.T) int64 {
	t.Helper()
	number := fmt.Sprintf("%s-%d", f.prefix, len(f.equipmentIDs)+1)
	var id int64
	if err := f.pool.QueryRow(context.Background(), `
		INSERT INTO equipment (inventory_number, serial_number, category_id, manufacturer, model, location_id)
		VALUES ($1, $1, $2, 'test', 'test', $3) RETURNING id
	`, number, f.categoryID, f.locationID).Scan(&id); err != nil {
		t.Fatalf("insert test equipment: %v", err)
	}
	f.equipmentIDs = append(f.equipmentIDs, id)
	return id
}

func TestRetireEquipmentPostgres(t *testing.T) {
	fixture := newRetirementFixture(t)
	ctx := context.Background()
	repository := equipment_postgres_repository.NewEquipmentRepository(fixture.pool)
	assignments := assignment_postgres_repository.NewAssignmentsRepository(fixture.pool)
	availableID := fixture.addEquipment(t)
	available, err := repository.GetEquipment(ctx, availableID)
	if err != nil {
		t.Fatalf("get available equipment: %v", err)
	}
	retired, err := repository.RetireEquipment(ctx, availableID)
	if err != nil || retired.Status != domain.EquipmentStatusRetired || retired.ID != availableID || !retired.UpdatedAt.After(available.UpdatedAt) || !retired.CreatedAt.Equal(available.CreatedAt) {
		t.Fatalf("RetireEquipment() = %+v, %v", retired, err)
	}
	_, err = repository.RetireEquipment(ctx, availableID)
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("repeat retirement error = %v", err)
	}
	_, err = assignments.SaveAssignment(ctx, domain.Assignment{
		EquipmentID: availableID,
		EmployeeID:  fixture.employeeID,
	})
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("issue retired equipment error = %v", err)
	}
	_, err = repository.RetireEquipment(ctx, -1)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("missing equipment retirement error = %v", err)
	}

	assignedID := fixture.addEquipment(t)
	issued, err := assignments.SaveAssignment(ctx, domain.Assignment{
		EquipmentID: assignedID,
		EmployeeID:  fixture.employeeID,
	})
	if err != nil {
		t.Fatalf("issue equipment: %v", err)
	}
	_, err = repository.RetireEquipment(ctx, assignedID)
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("retire assigned equipment error = %v", err)
	}
	if _, err := assignments.ReturnAssignment(ctx, assignedID); err != nil {
		t.Fatalf("return equipment: %v", err)
	}
	retired, err = repository.RetireEquipment(ctx, assignedID)
	if err != nil || retired.Status != domain.EquipmentStatusRetired {
		t.Fatalf("retire returned equipment = %+v, %v", retired, err)
	}
	history, err := assignments.GetEquipmentAssignments(ctx, assignedID)
	if err != nil || len(history) != 1 || history[0].ID != issued.ID || history[0].ReturnedAt == nil {
		t.Fatalf("history after retirement = %+v, %v", history, err)
	}

	inconsistentID := fixture.addEquipment(t)
	if _, err := fixture.pool.Exec(ctx, "INSERT INTO assignments (equipment_id, employee_id) VALUES ($1, $2)", inconsistentID, fixture.employeeID); err != nil {
		t.Fatalf("insert active assignment: %v", err)
	}
	_, err = repository.RetireEquipment(ctx, inconsistentID)
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("retire equipment with active assignment error = %v", err)
	}
	inconsistent, err := repository.GetEquipment(ctx, inconsistentID)
	if err != nil || inconsistent.Status != domain.EquipmentStatusInStock {
		t.Fatalf("equipment after conflict = %+v, %v", inconsistent, err)
	}
}

func TestRetirementAndIssueRacePostgres(t *testing.T) {
	fixture := newRetirementFixture(t)
	ctx := context.Background()
	repository := equipment_postgres_repository.NewEquipmentRepository(fixture.pool)
	assignments := assignment_postgres_repository.NewAssignmentsRepository(fixture.pool)
	equipmentID := fixture.addEquipment(t)
	start := make(chan struct{})
	var wait sync.WaitGroup
	var retireErr, issueErr error
	wait.Add(2)
	go func() {
		defer wait.Done()
		<-start
		_, retireErr = repository.RetireEquipment(ctx, equipmentID)
	}()
	go func() {
		defer wait.Done()
		<-start
		_, issueErr = assignments.SaveAssignment(ctx, domain.Assignment{
			EquipmentID: equipmentID,
			EmployeeID:  fixture.employeeID,
		})
	}()
	close(start)
	wait.Wait()
	if !((retireErr == nil && errors.Is(issueErr, core_errors.ErrConflict)) ||
		(issueErr == nil && errors.Is(retireErr, core_errors.ErrConflict))) {
		t.Fatalf("retirement error = %v, issue error = %v", retireErr, issueErr)
	}
	equipment, err := repository.GetEquipment(ctx, equipmentID)
	if err != nil {
		t.Fatalf("get equipment after race: %v", err)
	}
	var activeCount int
	if err := fixture.pool.QueryRow(ctx, "SELECT COUNT(*) FROM assignments WHERE equipment_id = $1 AND returned_at IS NULL", equipmentID).Scan(&activeCount); err != nil {
		t.Fatalf("count active assignments: %v", err)
	}
	if retireErr == nil && (equipment.Status != domain.EquipmentStatusRetired || activeCount != 0) ||
		issueErr == nil && (equipment.Status != domain.EquipmentStatusAssigned || activeCount != 1) {
		t.Fatalf("equipment status = %q, active assignments = %d", equipment.Status, activeCount)
	}
}
