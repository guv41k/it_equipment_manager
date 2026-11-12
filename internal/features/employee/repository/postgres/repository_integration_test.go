package employee_postgres_repository

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
	employee_service "gitlab.com/guv41k/it_equipment_manager/internal/features/employee/service"
)

var _ employee_service.EmployeesRepository = (*EmployeesRepository)(nil)

func TestEmployeesRepositoryPostgres(t *testing.T) {
	dsn := os.Getenv("EMPLOYEE_TEST_DSN")
	if dsn == "" {
		t.Skip("set EMPLOYEE_TEST_DSN to test against a migrated PostgreSQL database")
	}
	address, err := url.Parse(dsn)
	if err != nil || address.Scheme != "postgres" || address.User == nil || address.Port() == "" {
		t.Fatal("EMPLOYEE_TEST_DSN must be a postgres URL with user and port")
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

	repository := NewEmployeesRepository(pool)
	prefix := fmt.Sprintf("employee-test-%d", time.Now().UnixNano())
	var employeeIDs []int64
	var categoryID, locationID, equipmentID, assignmentID int64
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if assignmentID != 0 {
			if _, err := pool.Exec(cleanupCtx, "DELETE FROM assignments WHERE id = $1", assignmentID); err != nil {
				t.Errorf("delete test assignment: %v", err)
			}
		}
		if equipmentID != 0 {
			if _, err := pool.Exec(cleanupCtx, "DELETE FROM equipment WHERE id = $1", equipmentID); err != nil {
				t.Errorf("delete test equipment: %v", err)
			}
		}
		for _, id := range employeeIDs {
			if _, err := pool.Exec(cleanupCtx, "DELETE FROM employees WHERE id = $1", id); err != nil {
				t.Errorf("delete test employee %d: %v", id, err)
			}
		}
		if categoryID != 0 {
			if _, err := pool.Exec(cleanupCtx, "DELETE FROM equipment_categories WHERE id = $1", categoryID); err != nil {
				t.Errorf("delete test category: %v", err)
			}
		}
		if locationID != 0 {
			if _, err := pool.Exec(cleanupCtx, "DELETE FROM locations WHERE id = $1", locationID); err != nil {
				t.Errorf("delete test location: %v", err)
			}
		}
	})

	first, err := repository.SaveEmployee(ctx, domain.Employee{
		FullName:   "Иван Иванов",
		Department: "ИТ",
		IsActive:   true,
	})
	if err != nil || first.ID <= 0 || !first.IsActive {
		t.Fatalf("SaveEmployee() = %+v, %v", first, err)
	}
	employeeIDs = append(employeeIDs, first.ID)
	second, err := repository.SaveEmployee(ctx, domain.Employee{
		FullName:   "Иван Иванов",
		Department: "Бухгалтерия",
		IsActive:   true,
	})
	if err != nil || second.ID <= first.ID {
		t.Fatalf("second SaveEmployee() = %+v, %v", second, err)
	}
	employeeIDs = append(employeeIDs, second.ID)

	got, err := repository.GetEmployee(ctx, first.ID)
	if err != nil || got != first {
		t.Fatalf("GetEmployee() = %+v, %v", got, err)
	}
	list, err := repository.GetEmployees(ctx)
	if err != nil || !containsOrderedEmployees(list, first.ID, second.ID) {
		t.Fatalf("GetEmployees() = %+v, %v", list, err)
	}
	hasAssignments, err := repository.HasActiveAssignments(ctx, first.ID)
	if err != nil || hasAssignments {
		t.Fatalf("HasActiveAssignments() = %v, %v", hasAssignments, err)
	}

	first.Department = "Отдел разработки"
	first.IsActive = false
	updated, err := repository.UpdateEmployee(ctx, first)
	if err != nil || updated != first {
		t.Fatalf("UpdateEmployee() = %+v, %v", updated, err)
	}
	first.IsActive = true
	updated, err = repository.UpdateEmployee(ctx, first)
	if err != nil || updated != first {
		t.Fatalf("reactivate employee = %+v, %v", updated, err)
	}

	if err := pool.QueryRow(ctx, "INSERT INTO equipment_categories (name) VALUES ($1) RETURNING id", prefix).Scan(&categoryID); err != nil {
		t.Fatalf("insert test category: %v", err)
	}
	if err := pool.QueryRow(ctx, "INSERT INTO locations (name) VALUES ($1) RETURNING id", prefix).Scan(&locationID); err != nil {
		t.Fatalf("insert test location: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO equipment (inventory_number, serial_number, category_id, manufacturer, model, status, location_id)
		VALUES ($1, $2, $3, 'test', 'test', 'assigned', $4)
		RETURNING id
	`, prefix, prefix, categoryID, locationID).Scan(&equipmentID); err != nil {
		t.Fatalf("insert test equipment: %v", err)
	}
	if err := pool.QueryRow(ctx, "INSERT INTO assignments (equipment_id, employee_id) VALUES ($1, $2) RETURNING id", equipmentID, first.ID).Scan(&assignmentID); err != nil {
		t.Fatalf("insert test assignment: %v", err)
	}
	hasAssignments, err = repository.HasActiveAssignments(ctx, first.ID)
	if err != nil || !hasAssignments {
		t.Fatalf("HasActiveAssignments() = %v, %v", hasAssignments, err)
	}
	first.IsActive = false
	_, err = repository.UpdateEmployee(ctx, first)
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("deactivate assigned employee error = %v, want conflict", err)
	}
	got, err = repository.GetEmployee(ctx, first.ID)
	if err != nil || !got.IsActive {
		t.Fatalf("employee changed after conflict = %+v, %v", got, err)
	}
	if _, err := pool.Exec(ctx, "UPDATE assignments SET returned_at = NOW() WHERE id = $1", assignmentID); err != nil {
		t.Fatalf("return test assignment: %v", err)
	}
	hasAssignments, err = repository.HasActiveAssignments(ctx, first.ID)
	if err != nil || hasAssignments {
		t.Fatalf("HasActiveAssignments() after return = %v, %v", hasAssignments, err)
	}
	updated, err = repository.UpdateEmployee(ctx, first)
	if err != nil || updated.IsActive {
		t.Fatalf("deactivate returned employee = %+v, %v", updated, err)
	}

	_, err = repository.GetEmployee(ctx, -1)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("missing GetEmployee() error = %v, want not found", err)
	}
	_, err = repository.UpdateEmployee(ctx, domain.Employee{
		ID:         -1,
		FullName:   "Иван Иванов",
		Department: "ИТ",
		IsActive:   true,
	})
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("missing UpdateEmployee() error = %v, want not found", err)
	}
}

func containsOrderedEmployees(employees []domain.Employee, firstID, secondID int64) bool {
	firstIndex, secondIndex := -1, -1
	for i, employee := range employees {
		if employee.ID == firstID {
			firstIndex = i
		}
		if employee.ID == secondID {
			secondIndex = i
		}
	}
	return firstIndex >= 0 && secondIndex > firstIndex
}
