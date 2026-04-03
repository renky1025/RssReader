package api

import (
	"context"
	"net/http"
	"strings"

	"rssreader/internal/auth"
)

type contextKey string

const userContextKey contextKey = "user"

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeError(w, http.StatusUnauthorized, "invalid authorization header")
			return
		}

		claims, err := auth.ValidateToken(parts[1], s.config.JWTSecret)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromContext(r *http.Request) *auth.Claims {
	claims, ok := r.Context().Value(userContextKey).(*auth.Claims)
	if !ok {
		return nil
	}
	return claims
}

// AdminMiddleware checks if the user is an admin
func (s *Server) AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := getUserFromContext(r)
		if claims == nil || !claims.IsAdmin {
			writeError(w, http.StatusForbidden, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
