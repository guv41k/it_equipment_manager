package core_http_request

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

var requestValidator = validator.New()

type validatable interface {
	Validate() error
}

func DecodeAndValidateRequest(r *http.Request, dest any) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("decode JSON: %v: %w", err, core_errors.ErrInvalidArgument)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("request body must contain exactly one JSON value: %w", core_errors.ErrInvalidArgument)
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
