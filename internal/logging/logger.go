package logging

import (
	"log/slog"
	"os"
)

var (
	logger *slog.Logger
)

func GetLogger() *slog.Logger {
	return logger
}

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: false}))
	slog.SetDefault(logger)
}
