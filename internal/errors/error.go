package errors

import (
	"fmt"
	"net/http"
)

// Error represents a structured API error.
type Error struct {
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	StatusCode int            `json:"-"`
	Details    map[string]any `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// New creates a new Error.
func New(code, message string, statusCode int) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WithDetails adds details to the error.
func (e *Error) WithDetails(details map[string]any) *Error {
	e.Details = details
	return e
}

// Common error constructors
func BadRequest(message string) *Error {
	return New("BAD_REQUEST", message, http.StatusBadRequest)
}

func Unauthorized(message string) *Error {
	return New("UNAUTHORIZED", message, http.StatusUnauthorized)
}

func Forbidden(message string) *Error {
	return New("FORBIDDEN", message, http.StatusForbidden)
}

func NotFound(message string) *Error {
	return New("NOT_FOUND", message, http.StatusNotFound)
}

func Conflict(message string) *Error {
	return New("CONFLICT", message, http.StatusConflict)
}

func Internal(message string) *Error {
	return New("INTERNAL_ERROR", message, http.StatusInternalServerError)
}

func ServiceUnavailable(message string) *Error {
	return New("SERVICE_UNAVAILABLE", message, http.StatusServiceUnavailable)
}

// IsNotFound checks if the error is a not found error.
func IsNotFound(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.StatusCode == http.StatusNotFound
	}
	return false
}

// IsBadRequest checks if the error is a bad request error.
func IsBadRequest(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.StatusCode == http.StatusBadRequest
	}
	return false
}

// IsAlreadyExists checks if the error is a conflict error (already exists).
func IsAlreadyExists(err error) bool {
	return IsConflict(err)
}

// IsConflict checks if the error is a conflict error.
func IsConflict(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.StatusCode == http.StatusConflict
	}
	return false
}

// IsUnauthorized checks if the error is an unauthorized error.
func IsUnauthorized(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.StatusCode == http.StatusUnauthorized
	}
	return false
}

// FromError converts a standard error to an Error.
func FromError(err error) *Error {
	if e, ok := err.(*Error); ok {
		return e
	}
	return Internal(err.Error())
}
