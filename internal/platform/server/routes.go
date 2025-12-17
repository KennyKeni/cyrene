package server

import (
	"net/http"
	"strings"
)

func TrailingSlashMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && !strings.HasSuffix(r.URL.Path, "/") {
			r.URL.Path = r.URL.Path + "/"
		}
		next.ServeHTTP(w, r)
	})
}

func APIKeyMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if apiKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			path := r.URL.Path
			if path == "/" || path == "/health" || path == "/health/" ||
				strings.HasPrefix(path, "/swagger") {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("X-API-Key")
			if key == "" {
				auth := r.Header.Get("Authorization")
				if strings.HasPrefix(auth, "Bearer ") {
					key = strings.TrimPrefix(auth, "Bearer ")
				}
			}

			if key != apiKey {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware wraps a handler with CORS headers.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-API-Key")
		w.Header().Set("Access-Control-Allow-Credentials", "false")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
