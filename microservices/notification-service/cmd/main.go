package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/microservices/shared/config"
	"github.com/microservices/shared/logger"
	"github.com/microservices/shared/messaging"
	"github.com/microservices/notification-service/internal/consumer"
)

func main() {
	log := logger.New("notification-service")

	mq, err := messaging.Connect(config.LoadRabbitMQConfig().URL(), log)
	if err != nil {
		log.Error("connect to rabbitmq", "error", err)
		os.Exit(1)
	}
	defer mq.Close()

	c := consumer.New(mq, log)

	if err := c.Setup(); err != nil {
		log.Error("setup consumer", "error", err)
		os.Exit(1)
	}

	if err := c.Start(); err != nil {
		log.Error("start consumer", "error", err)
		os.Exit(1)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("notification-service stopped")
}
