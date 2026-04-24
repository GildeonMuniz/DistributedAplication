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
	"github.com/microservices/order-service/internal/events"
	"github.com/microservices/order-service/internal/handler"
	"github.com/microservices/order-service/internal/repository"
	"github.com/microservices/order-service/internal/service"
)

func main() {
	log := logger.New("order-service")

	db, err := database.Connect(config.LoadDatabaseConfig())
	if err != nil {
		log.Error("connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	mq, err := messaging.Connect(config.LoadRabbitMQConfig().URL(), log)
	if err != nil {
		log.Error("connect to rabbitmq", "error", err)
		os.Exit(1)
	}
	defer mq.Close()

	publisher := events.NewPublisher(mq)
	if err := publisher.Setup(); err != nil {
		log.Error("setup publisher", "error", err)
		os.Exit(1)
	}

	repo := repository.NewOrderRepository(db)
	svc := service.NewOrderService(repo, publisher, log)
	h := handler.NewOrderHandler(svc)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger(log))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "order-service"})
	})

	h.RegisterRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Info("order-service started", "port", port)
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
	log.Info("order-service stopped")
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
