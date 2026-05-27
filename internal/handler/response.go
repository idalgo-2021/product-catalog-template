package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/idalgo-2021/product-catalog-backend/internal/domain"

	"go.uber.org/zap"
)

type Response struct {
	Data  any        `json:"data,omitempty"`
	Meta  any        `json:"meta,omitempty"`
	Error *ErrorInfo `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func Success(w http.ResponseWriter, status int, data any) {
	JSON(w, status, Response{Data: data})
}

func SuccessWithMeta(w http.ResponseWriter, status int, data any, meta *Meta) {
	JSON(w, status, Response{Data: data, Meta: meta})
}

func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, Response{Error: &ErrorInfo{Code: code, Message: message}})
}

// HandleError maps domain errors to appropriate HTTP status codes and writes a JSON error response.
func HandleError(w http.ResponseWriter, logger *zap.Logger, msg string, err error) {
	logger.Error(msg, zap.Error(err))

	switch {
	case errors.Is(err, domain.ErrNotFound):
		Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrInvalidInput):
		Error(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
	case errors.Is(err, domain.ErrForbidden):
		Error(w, http.StatusForbidden, "FORBIDDEN", "Access denied")
	case errors.Is(err, domain.ErrUnauthorized):
		Error(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	case errors.Is(err, domain.ErrAlreadyExists):
		Error(w, http.StatusConflict, "ALREADY_EXISTS", err.Error())
	default:
		Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}
