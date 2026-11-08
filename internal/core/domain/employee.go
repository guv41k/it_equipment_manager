package domain

import (
	"fmt"
	"strings"
	"unicode/utf8"

	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

type Employee struct {
	ID         int64
	FullName   string
	Department string
	IsActive   bool
}

func (e Employee) Validate() error {
	if strings.TrimSpace(e.FullName) == "" || utf8.RuneCountInString(e.FullName) > 150 {
		return fmt.Errorf("full name must contain 1 to 150 characters: %w", core_errors.ErrInvalidArgument)
	}
	if strings.TrimSpace(e.Department) == "" || utf8.RuneCountInString(e.Department) > 100 {
		return fmt.Errorf("department must contain 1 to 100 characters: %w", core_errors.ErrInvalidArgument)
	}
	return nil
}
