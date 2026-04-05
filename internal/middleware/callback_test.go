package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCallbackMiddleware_CallbackSuccess(t *testing.T) {
	cfg := &CallbackConfig{
		Secure:      false,
		TTL:         3600,
		CookieName:  "session_id",
		FrontendURL: "https://example.com",
		Domain:      "",
		SameSite:    "lax",
	}

	// Create a handler that returns JSON response
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"sessionId": "test-session-123",
			"userId":    "user-456",
		})
	})

	middleware := CallbackMiddleware(cfg)
	handler := middleware(next)

	req := httptest.NewRequest("GET", "/auth/callback?code=test&state=test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should redirect to frontend
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "https://example.com", rec.Header().Get("Location"))

	// Should set session cookie
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session_id" {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie, "Session cookie should be set")
	assert.Equal(t, "test-session-123", sessionCookie.Value)
	assert.Equal(t, "/", sessionCookie.Path)
	assert.Equal(t, true, sessionCookie.HttpOnly)
	assert.Equal(t, false, sessionCookie.Secure)
	assert.Equal(t, 3600, sessionCookie.MaxAge)
	assert.Equal(t, http.SameSiteLaxMode, sessionCookie.SameSite)
}

func TestCallbackMiddleware_CallbackError(t *testing.T) {
	cfg := &CallbackConfig{
		Secure:      false,
		TTL:         3600,
		CookieName:  "session_id",
		FrontendURL: "https://example.com",
		Domain:      "",
		SameSite:    "lax",
	}

	// Create a handler that returns error response
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"type":    "invalid_request",
			"message": "Invalid code",
		})
	})

	middleware := CallbackMiddleware(cfg)
	handler := middleware(next)

	req := httptest.NewRequest("GET", "/auth/callback?error=invalid", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should return error response, NOT redirect
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "", rec.Header().Get("Location"))

	// Verify error response body
	var resp map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_request", resp["type"])
}

func TestCallbackMiddleware_Logout(t *testing.T) {
	cfg := &CallbackConfig{
		Secure:      false,
		TTL:         3600,
		CookieName:  "session_id",
		FrontendURL: "https://example.com",
		Domain:      "",
		SameSite:    "lax",
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	middleware := CallbackMiddleware(cfg)
	handler := middleware(next)

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should clear session cookie
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session_id" {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie, "Session cookie should be set")
	assert.Equal(t, "", sessionCookie.Value)
	assert.Equal(t, -1, sessionCookie.MaxAge) // Cleared
}

func TestCallbackMiddleware_WithDomain(t *testing.T) {
	cfg := &CallbackConfig{
		Secure:      false,
		TTL:         3600,
		CookieName:  "session_id",
		FrontendURL: "https://example.com",
		Domain:      ".example.com",
		SameSite:    "lax",
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"sessionId": "test-session-123",
		})
	})

	middleware := CallbackMiddleware(cfg)
	handler := middleware(next)

	req := httptest.NewRequest("GET", "/auth/callback?code=test&state=test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session_id" {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie)
	assert.Equal(t, "example.com", sessionCookie.Domain)
}

func TestParseSameSite(t *testing.T) {
	tests := []struct {
		input    string
		expected http.SameSite
	}{
		{"lax", http.SameSiteLaxMode},
		{"Lax", http.SameSiteLaxMode},
		{"LAX", http.SameSiteLaxMode},
		{"strict", http.SameSiteStrictMode},
		{"Strict", http.SameSiteStrictMode},
		{"none", http.SameSiteNoneMode},
		{"", http.SameSiteNoneMode},
		{"invalid", http.SameSiteLaxMode}, // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseSameSite(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
