package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Ashoke15/AuthX/internal/auth"
)

type contextKey string

const userIdContextKey contextKey = "userId"

func RequireAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				writeUnauthorized(w, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)

			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeUnauthorized(w, "authorization header must be: Bearer <token>")
				return
			}

			claims, err := auth.ParseToken(parts[1], jwtSecret)
			if err != nil {
				writeUnauthorized(w, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIdContextKey, claims.UserId)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIdFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIdContextKey).(string)
	return id, ok
}

func writeUnauthorized(w http.ResponseWriter, massege string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": massege})
}
