package core_http_request

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

var requestValidator = validator.New()

type validatable interface {
	Validate() error
}

func DecodeAndValidateRequest(r *http.Request, dest any) error {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		return fmt.Errorf("decode JSON: %v: %w", err, core_errors.ErrInvalidArgument)
	}
	var err error
	if value, ok := dest.(validatable); ok {
		err = value.Validate()
	} else {
		err = requestValidator.Struct(dest)
	}
	if err != nil {
		return fmt.Errorf("validate request: %v: %w", err, core_errors.ErrInvalidArgument)
	}
	return nil
}
