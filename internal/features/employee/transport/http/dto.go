package employee_transport_http

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

type CreateEmployeeRequest struct {
	FullName   string `json:"full_name"`
	Department string `json:"department"`
}

func (request CreateEmployeeRequest) Validate() error {
	return validateEmployeeFields(request.FullName, request.Department)
}

type UpdateEmployeeRequest struct {
	FullName   string `json:"full_name"`
	Department string `json:"department"`
	IsActive   *bool  `json:"is_active"`
}

func (request UpdateEmployeeRequest) Validate() error {
	if err := validateEmployeeFields(request.FullName, request.Department); err != nil {
		return err
	}
	if request.IsActive == nil {
		return fmt.Errorf("is_active is required")
	}
	return nil
}

func validateEmployeeFields(fullName, department string) error {
	if strings.TrimSpace(fullName) == "" || utf8.RuneCountInString(fullName) > 150 {
		return fmt.Errorf("full_name must contain 1 to 150 characters")
	}
	if strings.TrimSpace(department) == "" || utf8.RuneCountInString(department) > 100 {
		return fmt.Errorf("department must contain 1 to 100 characters")
	}
	return nil
}

type EmployeeResponse struct {
	ID         int64  `json:"id"`
	FullName   string `json:"full_name"`
	Department string `json:"department"`
	IsActive   bool   `json:"is_active"`
}

func employeeDTOFromDomain(employee domain.Employee) EmployeeResponse {
	return EmployeeResponse{
		ID:         employee.ID,
		FullName:   employee.FullName,
		Department: employee.Department,
		IsActive:   employee.IsActive,
	}
}
