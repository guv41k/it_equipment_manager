package core_http_request

import (
	"fmt"
	"net/http"
	"strconv"

	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

func GetIntQueryParam(r *http.Request, key string) (*int, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf("%s must be an integer: %w", key, core_errors.ErrInvalidArgument)
	}
	return &parsed, nil
}

func GetInt64QueryParam(r *http.Request, key string) (*int64, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%s must be an integer: %w", key, core_errors.ErrInvalidArgument)
	}
	return &parsed, nil
}
