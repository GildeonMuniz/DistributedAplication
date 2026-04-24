package config

import (
	"fmt"
	"os"
	"strconv"
)

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"sqlserver://%s:%s@%s:%d?database=%s&encrypt=disable",
		c.User, c.Password, c.Host, c.Port, c.Database,
	)
}

type RabbitMQConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	VHost    string
}

func (c RabbitMQConfig) URL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", c.User, c.Password, c.Host, c.Port, c.VHost)
}

type JWTConfig struct {
	Secret     string
	Expiration int // hours
}

func LoadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 1433),
		User:     getEnv("DB_USER", "sa"),
		Password: getEnv("DB_PASSWORD", "YourStrong@Passw0rd"),
		Database: getEnv("DB_NAME", "microservices"),
	}
}

func LoadRabbitMQConfig() RabbitMQConfig {
	return RabbitMQConfig{
		Host:     getEnv("RABBITMQ_HOST", "localhost"),
		Port:     getEnvInt("RABBITMQ_PORT", 5672),
		User:     getEnv("RABBITMQ_USER", "guest"),
		Password: getEnv("RABBITMQ_PASSWORD", "guest"),
		VHost:    getEnv("RABBITMQ_VHOST", "/"),
	}
}

func LoadJWTConfig() JWTConfig {
	return JWTConfig{
		Secret:     getEnv("JWT_SECRET", "super-secret-key-change-in-production"),
		Expiration: getEnvInt("JWT_EXPIRATION_HOURS", 24),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
