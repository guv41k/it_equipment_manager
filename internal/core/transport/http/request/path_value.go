package core_http_request

import (
	"fmt"
	"net/http"
	"strconv"

	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
)

func GetInt64PathValue(r *http.Request, key string) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue(key), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer: %w", key, core_errors.ErrInvalidArgument)
	}
	return id, nil
}
