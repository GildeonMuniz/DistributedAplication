package middleware

import (
	"net/http"

	"github.com/microservices/shared/config"
	sharedmw "github.com/microservices/shared/middleware"
)

func JWTAuth(cfg config.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := sharedmw.ExtractBearerToken(r)
			if token == "" {
				http.Error(w, `{"message":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			claims, err := sharedmw.ValidateToken(token, cfg.Secret)
			if err != nil {
				http.Error(w, `{"message":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			r.Header.Set("X-User-ID", claims.UserID)
			r.Header.Set("X-User-Email", claims.Email)
			r.Header.Set("X-User-Role", claims.Role)

			next.ServeHTTP(w, r)
		})
	}
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
