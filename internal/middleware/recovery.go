package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// RecoveryMiddleware creates a middleware that recovers from panics.
func RecoveryMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rv := recover(); rv != nil {
					fmt.Printf("PANIC recovered: %v\n", rv)
					fmt.Printf("Stack: %s\n", debug.Stack())
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
