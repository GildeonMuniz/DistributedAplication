package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Detail)
}

func New(code int, message, detail string) *AppError {
	return &AppError{Code: code, Message: message, Detail: detail}
}

var (
	ErrNotFound     = New(http.StatusNotFound, "resource not found", "")
	ErrBadRequest   = New(http.StatusBadRequest, "invalid request", "")
	ErrUnauthorized = New(http.StatusUnauthorized, "unauthorized", "")
	ErrInternal     = New(http.StatusInternalServerError, "internal server error", "")
	ErrConflict     = New(http.StatusConflict, "resource already exists", "")
)

func IsNotFound(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == http.StatusNotFound
}

func Wrap(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}
