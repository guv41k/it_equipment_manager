package domain

import (
	"fmt"
	"strings"
	"unicode/utf8"

	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

func validateCatalogFields(name string, description *string) error {
	if strings.TrimSpace(name) == "" || utf8.RuneCountInString(name) > 100 {
		return fmt.Errorf("name must contain 1 to 100 characters: %w", core_errors.ErrInvalidArgument)
	}
	if description != nil && utf8.RuneCountInString(*description) > 500 {
		return fmt.Errorf("description must have at most 500 characters: %w", core_errors.ErrInvalidArgument)
	}
	return nil
}
