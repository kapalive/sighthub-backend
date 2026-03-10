package auth

import (
	"encoding/json"
	"net/http"

	"github.com/redis/go-redis/v9"
)

// JWTMiddleware проверяет Bearer токен и Redis blacklist.
// При успехе устанавливает username в context.
func JWTMiddleware(jwtSecret string, rdb *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := ExtractTokenFromHeader(r)
			if tokenStr == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authorization token required"})
				return
			}

			claims, err := ParseToken(tokenStr, jwtSecret)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
				return
			}

			// Проверяем Redis blacklist
			ctx := r.Context()
			if rdb != nil {
				exists, err := rdb.Exists(ctx, claims.ID).Result()
				if err == nil && exists > 0 {
					writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "token has been revoked"})
					return
				}
			}

			next.ServeHTTP(w, r.WithContext(WithUsername(ctx, claims.Username)))
		})
	}
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
