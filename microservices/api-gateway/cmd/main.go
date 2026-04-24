package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/microservices/shared/config"
	"github.com/microservices/shared/logger"
	gw "github.com/microservices/api-gateway/internal/middleware"
	"github.com/microservices/api-gateway/internal/proxy"
)

func main() {
	log := logger.New("api-gateway")
	jwtCfg := config.LoadJWTConfig()

	services := map[string]string{
		"users":  getEnv("USER_SERVICE_URL", "http://user-service:8081"),
		"orders": getEnv("ORDER_SERVICE_URL", "http://order-service:8082"),
	}

	p, err := proxy.New(services)
	if err != nil {
		log.Error("create proxy", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "api-gateway"})
	})

	// Public routes (no auth)
	mux.Handle("/api/v1/auth/", p.Handler("users"))
	mux.Handle("/api/v1/users", withMethod(http.MethodPost, p.Handler("users"))) // register

	// Protected routes
	auth := gw.JWTAuth(jwtCfg)
	mux.Handle("/api/v1/users/", auth(p.Handler("users")))
	mux.Handle("/api/v1/orders", auth(p.Handler("orders")))
	mux.Handle("/api/v1/orders/", auth(p.Handler("orders")))

	handler := gw.CORS(requestLogger(log)(mux))

	port := getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("api-gateway started", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Info("api-gateway stopped")
}

func withMethod(method string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == method {
			next.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})
}

func requestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, code: http.StatusOK}
			next.ServeHTTP(rw, r)
			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.code,
				"duration", time.Since(start),
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	code int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.code = code
	rw.ResponseWriter.WriteHeader(code)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
