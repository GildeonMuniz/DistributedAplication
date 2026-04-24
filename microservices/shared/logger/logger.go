package logger

import (
	"log/slog"
	"os"
)

func New(service string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if os.Getenv("LOG_LEVEL") == "debug" {
		opts.Level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler).With("service", service)
}
