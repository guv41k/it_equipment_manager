package core_http_response

import (
	"encoding/json"
	"errors"
	"net/http"

	core_errors "gitlab.com/guv41k/it_equipment_manager/internal/core/errors"
	core_logger "gitlab.com/guv41k/it_equipment_manager/internal/core/logger"
	"go.uber.org/zap"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type HTTPResponseHandler struct {
	log *core_logger.Logger
	rw  http.ResponseWriter
}

func NewHTTPResponseHandler(log *core_logger.Logger, rw http.ResponseWriter) *HTTPResponseHandler {
	return &HTTPResponseHandler{
		log: log,
		rw:  rw,
	}
}

func (h *HTTPResponseHandler) JSONResponse(body any, status int) {
	h.rw.Header().Set("Content-Type", "application/json")
	h.rw.WriteHeader(status)
	if err := json.NewEncoder(h.rw).Encode(body); err != nil {
		h.log.Error("write HTTP response", zap.Error(err))
	}
}

func (h *HTTPResponseHandler) ErrorResponse(err error, operation string) {
	status := http.StatusInternalServerError
	response := ErrorResponse{
		Error:   "internal_error",
		Message: "internal server error",
	}
	switch {
	case errors.Is(err, core_errors.ErrInvalidArgument):
		h.log.Warn(operation, zap.Error(err))
		status = http.StatusBadRequest
		response.Error = "invalid_request"
		response.Message = err.Error()
	case errors.Is(err, core_errors.ErrNotFound):
		h.log.Debug(operation, zap.Error(err))
		status = http.StatusNotFound
		response.Error = "not_found"
		response.Message = "record not found"
	case errors.Is(err, core_errors.ErrConflict):
		h.log.Warn(operation, zap.Error(err))
		status = http.StatusConflict
		response.Error = "conflict"
		response.Message = "operation conflicts with current state"
	default:
		h.log.Error(operation, zap.Error(err))
	}
	h.JSONResponse(response, status)
}
