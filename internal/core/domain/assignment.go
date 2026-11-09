package domain

import (
	"fmt"
	"time"

	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

type Assignment struct {
	ID          int64
	EquipmentID int64
	EmployeeID  int64
	IssuedAt    time.Time
	ReturnedAt  *time.Time
}

func (a Assignment) Validate() error {
	if a.EquipmentID <= 0 {
		return fmt.Errorf("equipment id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	if a.EmployeeID <= 0 {
		return fmt.Errorf("employee id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	if a.ReturnedAt != nil && (a.IssuedAt.IsZero() || a.ReturnedAt.Before(a.IssuedAt)) {
		return fmt.Errorf("returned at must not precede issued at: %w", core_errors.ErrInvalidArgument)
	}
	return nil
}
