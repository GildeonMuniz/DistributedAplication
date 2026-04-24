package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/microservices/shared/config"
	"github.com/microservices/shared/database"
	"github.com/microservices/shared/logger"
	"github.com/microservices/shared/messaging"
	"github.com/microservices/user-service/internal/events"
	"github.com/microservices/user-service/internal/handler"
	"github.com/microservices/user-service/internal/repository"
	"github.com/microservices/user-service/internal/service"
)

func main() {
	log := logger.New("user-service")

	dbCfg := config.LoadDatabaseConfig()
	db, err := database.Connect(dbCfg)
	if err != nil {
		log.Error("connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("database connected")

	mqCfg := config.LoadRabbitMQConfig()
	mq, err := messaging.Connect(mqCfg.URL(), log)
	if err != nil {
		log.Error("connect to rabbitmq", "error", err)
		os.Exit(1)
	}
	defer mq.Close()
	log.Info("rabbitmq connected")

	publisher := events.NewPublisher(mq)
	if err := publisher.Setup(); err != nil {
		log.Error("setup publisher", "error", err)
		os.Exit(1)
	}

	jwtCfg := config.LoadJWTConfig()
	repo := repository.NewUserRepository(db)
	svc := service.NewUserService(repo, publisher, jwtCfg, log)
	h := handler.NewUserHandler(svc)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger(log))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "user-service"})
	})

	h.RegisterRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("user-service started", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func requestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", time.Since(start),
		)
	}
}
