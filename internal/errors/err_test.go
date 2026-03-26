package errors

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	err := New("TEST_CODE", "test message", http.StatusBadRequest)

	assert.Equal(t, "TEST_CODE", err.Code)
	assert.Equal(t, "test message", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
	assert.Nil(t, err.Details)
}

func TestError_Error(t *testing.T) {
	err := New("TEST_CODE", "test message", http.StatusBadRequest)

	assert.Equal(t, "TEST_CODE: test message", err.Error())
}

func TestWithDetails(t *testing.T) {
	err := New("TEST_CODE", "test message", http.StatusBadRequest)
	details := map[string]interface{}{
		"field": "username",
		"value": "test",
	}

	result := err.WithDetails(details)

	assert.Equal(t, err, result)
	assert.Equal(t, details, err.Details)
}

func TestBadRequest(t *testing.T) {
	err := BadRequest("invalid request")

	assert.Equal(t, "BAD_REQUEST", err.Code)
	assert.Equal(t, "invalid request", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
}

func TestUnauthorized(t *testing.T) {
	err := Unauthorized("not authorized")

	assert.Equal(t, "UNAUTHORIZED", err.Code)
	assert.Equal(t, "not authorized", err.Message)
	assert.Equal(t, http.StatusUnauthorized, err.StatusCode)
}

func TestForbidden(t *testing.T) {
	err := Forbidden("access denied")

	assert.Equal(t, "FORBIDDEN", err.Code)
	assert.Equal(t, "access denied", err.Message)
	assert.Equal(t, http.StatusForbidden, err.StatusCode)
}

func TestNotFound(t *testing.T) {
	err := NotFound("resource not found")

	assert.Equal(t, "NOT_FOUND", err.Code)
	assert.Equal(t, "resource not found", err.Message)
	assert.Equal(t, http.StatusNotFound, err.StatusCode)
}

func TestConflict(t *testing.T) {
	err := Conflict("resource already exists")

	assert.Equal(t, "CONFLICT", err.Code)
	assert.Equal(t, "resource already exists", err.Message)
	assert.Equal(t, http.StatusConflict, err.StatusCode)
}

func TestInternal(t *testing.T) {
	err := Internal("internal server error")

	assert.Equal(t, "INTERNAL_ERROR", err.Code)
	assert.Equal(t, "internal server error", err.Message)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode)
}

func TestServiceUnavailable(t *testing.T) {
	err := ServiceUnavailable("service temporarily unavailable")

	assert.Equal(t, "SERVICE_UNAVAILABLE", err.Code)
	assert.Equal(t, "service temporarily unavailable", err.Message)
	assert.Equal(t, http.StatusServiceUnavailable, err.StatusCode)
}

func TestIsNotFound(t *testing.T) {
	t.Run("returns true for not found error", func(t *testing.T) {
		err := NotFound("resource not found")
		assert.True(t, IsNotFound(err))
	})

	t.Run("returns false for other errors", func(t *testing.T) {
		err := BadRequest("bad request")
		assert.False(t, IsNotFound(err))
	})

	t.Run("returns false for standard error", func(t *testing.T) {
		err := assert.AnError
		assert.False(t, IsNotFound(err))
	})
}

func TestIsConflict(t *testing.T) {
	t.Run("returns true for conflict error", func(t *testing.T) {
		err := Conflict("resource conflict")
		assert.True(t, IsConflict(err))
	})

	t.Run("returns false for other errors", func(t *testing.T) {
		err := BadRequest("bad request")
		assert.False(t, IsConflict(err))
	})
}

func TestIsUnauthorized(t *testing.T) {
	t.Run("returns true for unauthorized error", func(t *testing.T) {
		err := Unauthorized("not authorized")
		assert.True(t, IsUnauthorized(err))
	})

	t.Run("returns false for other errors", func(t *testing.T) {
		err := BadRequest("bad request")
		assert.False(t, IsUnauthorized(err))
	})
}

func TestFromError(t *testing.T) {
	t.Run("returns same error for Error type", func(t *testing.T) {
		original := NotFound("resource not found")
		result := FromError(original)

		assert.Equal(t, original, result)
	})

	t.Run("wraps standard error as internal error", func(t *testing.T) {
		original := assert.AnError
		result := FromError(original)

		assert.Equal(t, "INTERNAL_ERROR", result.Code)
		assert.Equal(t, original.Error(), result.Message)
		assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
	})
}
