package assignment_transport_http

import (
	"fmt"
	"time"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

type IssueEquipmentRequest struct {
	EmployeeID int64 `json:"employee_id"`
}

func (request IssueEquipmentRequest) Validate() error {
	if request.EmployeeID <= 0 {
		return fmt.Errorf("employee_id must be positive")
	}
	return nil
}

type AssignmentResponse struct {
	ID          int64      `json:"id"`
	EquipmentID int64      `json:"equipment_id"`
	EmployeeID  int64      `json:"employee_id"`
	IssuedAt    time.Time  `json:"issued_at"`
	ReturnedAt  *time.Time `json:"returned_at"`
}

func assignmentDTOFromDomain(assignment domain.Assignment) AssignmentResponse {
	return AssignmentResponse{
		ID:          assignment.ID,
		EquipmentID: assignment.EquipmentID,
		EmployeeID:  assignment.EmployeeID,
		IssuedAt:    assignment.IssuedAt,
		ReturnedAt:  assignment.ReturnedAt,
	}
}
