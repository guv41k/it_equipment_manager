package assignment_postgres_repository

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
	assignment_service "gitlab.com/guv41k/it_equipment_manager/internal/features/assignment/service"
	employee_postgres_repository "gitlab.com/guv41k/it_equipment_manager/internal/features/employee/repository/postgres"
)

var _ assignment_service.AssignmentsRepository = (*AssignmentsRepository)(nil)

type assignmentFixture struct {
	pool         *core_pgx_pool.Pool
	prefix       string
	categoryID   int64
	locationID   int64
	employeeIDs  []int64
	equipmentIDs []int64
}

func newAssignmentFixture(t *testing.T) *assignmentFixture {
	t.Helper()
	dsn := os.Getenv("ASSIGNMENT_TEST_DSN")
	if dsn == "" {
		t.Skip("set ASSIGNMENT_TEST_DSN to test against a migrated PostgreSQL database")
	}
	address, err := url.Parse(dsn)
	if err != nil || address.Scheme != "postgres" || address.User == nil || address.Port() == "" {
		t.Fatal("ASSIGNMENT_TEST_DSN must be a postgres URL with user and port")
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
	fixture := &assignmentFixture{
		pool:   pool,
		prefix: fmt.Sprintf("assignment-test-%d", time.Now().UnixNano()),
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
		for _, id := range fixture.employeeIDs {
			if _, err := pool.Exec(ctx, "DELETE FROM employees WHERE id = $1", id); err != nil {
				t.Errorf("delete test employee %d: %v", id, err)
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
	return fixture
}

func (f *assignmentFixture) addEmployee(t *testing.T, active bool) int64 {
	t.Helper()
	var id int64
	if err := f.pool.QueryRow(context.Background(), `
		INSERT INTO employees (full_name, department, is_active)
		VALUES ($1, 'ИТ', $2) RETURNING id
	`, f.prefix, active).Scan(&id); err != nil {
		t.Fatalf("insert test employee: %v", err)
	}
	f.employeeIDs = append(f.employeeIDs, id)
	return id
}

func (f *assignmentFixture) addEquipment(t *testing.T) int64 {
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

func TestAssignmentsRepositoryPostgres(t *testing.T) {
	fixture := newAssignmentFixture(t)
	ctx := context.Background()
	repository := NewAssignmentsRepository(fixture.pool)
	service := assignment_service.NewAssignmentsService(repository)
	equipmentID := fixture.addEquipment(t)
	firstEmployeeID := fixture.addEmployee(t, true)
	secondEmployeeID := fixture.addEmployee(t, true)
	inactiveEmployeeID := fixture.addEmployee(t, false)
	inactiveEquipmentID := fixture.addEquipment(t)
	_, err := repository.SaveAssignment(ctx, domain.Assignment{
		EquipmentID: inactiveEquipmentID,
		EmployeeID:  inactiveEmployeeID,
	})
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("issue to inactive employee error = %v", err)
	}

	equipment, err := repository.GetEquipment(ctx, equipmentID)
	if err != nil || equipment.Status != domain.EquipmentStatusInStock {
		t.Fatalf("GetEquipment() = %+v, %v", equipment, err)
	}
	employee, err := repository.GetEmployee(ctx, firstEmployeeID)
	if err != nil || !employee.IsActive {
		t.Fatalf("GetEmployee() = %+v, %v", employee, err)
	}
	hasActive, err := repository.HasActiveAssignment(ctx, equipmentID)
	if err != nil || hasActive {
		t.Fatalf("HasActiveAssignment() = %v, %v", hasActive, err)
	}
	history, err := service.GetEquipmentAssignments(ctx, equipmentID)
	if err != nil || len(history) != 0 {
		t.Fatalf("empty history = %+v, %v", history, err)
	}

	_, err = repository.SaveAssignment(ctx, domain.Assignment{
		EquipmentID: equipmentID,
		EmployeeID:  -1,
	})
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("issue to missing employee error = %v", err)
	}
	issued, err := service.IssueEquipment(ctx, equipmentID, firstEmployeeID)
	if err != nil || issued.ID <= 0 || issued.EquipmentID != equipmentID || issued.EmployeeID != firstEmployeeID || issued.IssuedAt.IsZero() || issued.ReturnedAt != nil {
		t.Fatalf("IssueEquipment() = %+v, %v", issued, err)
	}
	equipment, err = repository.GetEquipment(ctx, equipmentID)
	if err != nil || equipment.Status != domain.EquipmentStatusAssigned || equipment.UpdatedAt.Before(issued.IssuedAt) {
		t.Fatalf("equipment after issue = %+v, %v", equipment, err)
	}
	hasActive, err = repository.HasActiveAssignment(ctx, equipmentID)
	if err != nil || !hasActive {
		t.Fatalf("active assignment after issue = %v, %v", hasActive, err)
	}
	_, err = repository.SaveAssignment(ctx, domain.Assignment{
		EquipmentID: equipmentID,
		EmployeeID:  secondEmployeeID,
	})
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("duplicate issue error = %v", err)
	}

	returned, err := service.ReturnEquipment(ctx, equipmentID)
	if err != nil || returned.ID != issued.ID || returned.ReturnedAt == nil || returned.ReturnedAt.Before(issued.IssuedAt) {
		t.Fatalf("ReturnEquipment() = %+v, %v", returned, err)
	}
	equipment, err = repository.GetEquipment(ctx, equipmentID)
	if err != nil || equipment.Status != domain.EquipmentStatusInStock || equipment.UpdatedAt.Before(*returned.ReturnedAt) {
		t.Fatalf("equipment after return = %+v, %v", equipment, err)
	}
	_, err = repository.ReturnAssignment(ctx, equipmentID)
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("second return error = %v", err)
	}
	second, err := service.IssueEquipment(ctx, equipmentID, secondEmployeeID)
	if err != nil || second.ID == issued.ID || second.EmployeeID != secondEmployeeID || second.ReturnedAt != nil {
		t.Fatalf("second IssueEquipment() = %+v, %v", second, err)
	}
	history, err = service.GetEquipmentAssignments(ctx, equipmentID)
	if err != nil || len(history) != 2 || history[0].ID != issued.ID || history[0].ReturnedAt == nil || history[1].ID != second.ID || history[1].ReturnedAt != nil {
		t.Fatalf("assignment history = %+v, %v", history, err)
	}
}

func TestConcurrentAssignmentsPostgres(t *testing.T) {
	fixture := newAssignmentFixture(t)
	repository := NewAssignmentsRepository(fixture.pool)
	equipmentID := fixture.addEquipment(t)
	firstEmployeeID := fixture.addEmployee(t, true)
	secondEmployeeID := fixture.addEmployee(t, true)
	start := make(chan struct{})
	var wait sync.WaitGroup
	errorsByRequest := make([]error, 2)
	for i, employeeID := range []int64{firstEmployeeID, secondEmployeeID} {
		wait.Add(1)
		go func(i int, employeeID int64) {
			defer wait.Done()
			<-start
			_, errorsByRequest[i] = repository.SaveAssignment(context.Background(), domain.Assignment{
				EquipmentID: equipmentID,
				EmployeeID:  employeeID,
			})
		}(i, employeeID)
	}
	close(start)
	wait.Wait()
	if !((errorsByRequest[0] == nil && errors.Is(errorsByRequest[1], core_errors.ErrConflict)) ||
		(errorsByRequest[1] == nil && errors.Is(errorsByRequest[0], core_errors.ErrConflict))) {
		t.Fatalf("concurrent issue errors = %v", errorsByRequest)
	}
	var activeCount int
	if err := fixture.pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM assignments WHERE equipment_id = $1 AND returned_at IS NULL", equipmentID).Scan(&activeCount); err != nil || activeCount != 1 {
		t.Fatalf("active assignment count = %d, %v", activeCount, err)
	}
}

func TestIssueAndEmployeeDeactivationPostgres(t *testing.T) {
	fixture := newAssignmentFixture(t)
	repository := NewAssignmentsRepository(fixture.pool)
	employeesRepository := employee_postgres_repository.NewEmployeesRepository(fixture.pool)
	equipmentID := fixture.addEquipment(t)
	employeeID := fixture.addEmployee(t, true)
	start := make(chan struct{})
	var wait sync.WaitGroup
	var issueErr, deactivateErr error
	wait.Add(2)
	go func() {
		defer wait.Done()
		<-start
		_, issueErr = repository.SaveAssignment(context.Background(), domain.Assignment{
			EquipmentID: equipmentID,
			EmployeeID:  employeeID,
		})
	}()
	go func() {
		defer wait.Done()
		<-start
		_, deactivateErr = employeesRepository.UpdateEmployee(context.Background(), domain.Employee{
			ID:         employeeID,
			FullName:   fixture.prefix,
			Department: "ИТ",
			IsActive:   false,
		})
	}()
	close(start)
	wait.Wait()
	if !((issueErr == nil && errors.Is(deactivateErr, core_errors.ErrConflict)) ||
		(deactivateErr == nil && errors.Is(issueErr, core_errors.ErrConflict))) {
		t.Fatalf("issue error = %v, deactivate error = %v", issueErr, deactivateErr)
	}
	employee, err := repository.GetEmployee(context.Background(), employeeID)
	if err != nil {
		t.Fatalf("get employee after race: %v", err)
	}
	hasActive, err := repository.HasActiveAssignment(context.Background(), equipmentID)
	if err != nil || hasActive != employee.IsActive {
		t.Fatalf("employee active = %v, assignment active = %v, error = %v", employee.IsActive, hasActive, err)
	}
}
