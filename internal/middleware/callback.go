package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/starfrag-lab/retrowin-go/internal/config"
)

// CallbackConfig holds the configuration for the callback middleware.
type CallbackConfig struct {
	Secure      bool
	TTL         int
	CookieName  string
	FrontendURL string
	Domain      string
	SameSite    string
}

// CallbackMiddleware creates a middleware that handles OAuth callback and logout:
// - On callback: captures the response body to extract session ID and sets the cookie, then redirects to frontend.
// - On logout: clears the session cookie.
func CallbackMiddleware(cfg *CallbackConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		parsedSameSite := parseSameSite(cfg.SameSite)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Logout: clear session cookie before the handler runs
			if r.Method == http.MethodPost && r.URL.Path == "/auth/logout" {
				cookie := &http.Cookie{
					Name:     cfg.CookieName,
					Value:    "",
					Path:     "/",
					HttpOnly: true,
					Secure:   cfg.Secure,
					MaxAge:   -1,
					SameSite: parsedSameSite,
				}
				if cfg.Domain != "" {
					cookie.Domain = cfg.Domain
				}
				http.SetCookie(w, cookie)
				next.ServeHTTP(w, r)
				return
			}

			// Callback: capture response to extract session ID, set cookie, and redirect to frontend
			if r.Method == http.MethodGet && r.URL.Path == "/auth/callback" {
				rec := &responseRecorder{ResponseWriter: w}
				next.ServeHTTP(rec, r)

				// Only set cookie on successful callback (200 status) and redirect
				if rec.statusCode == http.StatusOK && len(rec.body) > 0 {
					var resp struct {
						SessionID string `json:"sessionId"`
					}
					if err := json.Unmarshal(rec.body, &resp); err == nil && resp.SessionID != "" {
						cookie := &http.Cookie{
							Name:     cfg.CookieName,
							Value:    resp.SessionID,
							Path:     "/",
							HttpOnly: true,
							Secure:   cfg.Secure,
							MaxAge:   cfg.TTL,
							SameSite: parsedSameSite,
						}
						if cfg.Domain != "" {
							cookie.Domain = cfg.Domain
						}
						http.SetCookie(w, cookie)
						// Redirect to frontend after successful login
						http.Redirect(w, r, cfg.FrontendURL, http.StatusFound)
						return
					}
				}
				// On error, return the original response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(rec.statusCode)
				_, _ = w.Write(rec.body)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// parseSameSite parses the SameSite configuration string.
func parseSameSite(sameSite string) http.SameSite {
	switch strings.ToLower(sameSite) {
	case "strict":
		return http.SameSiteStrictMode
	case "none", "":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

// responseRecorder captures the response body and status code.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}
	r.body = append(r.body, b...)
	// Don't write to the original ResponseWriter yet
	// We'll write it later after deciding whether to redirect
	return len(b), nil
}

// ProvideCallbackConfig provides the callback middleware configuration from the application config.
func ProvideCallbackConfig(cfg *config.Config) *CallbackConfig {
	return &CallbackConfig{
		Secure:      cfg.Auth.Session.Secure,
		TTL:         cfg.Auth.Session.TTL,
		CookieName:  cfg.Auth.Session.CookieName,
		FrontendURL: cfg.Auth.Session.FrontendURL,
		Domain:      cfg.Auth.Session.Domain,
		SameSite:    cfg.Auth.Session.SameSite,
	}
}
