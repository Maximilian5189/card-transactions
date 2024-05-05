package logger

import (
	"log/slog"
	"os"
)

type Logger struct{}

func NewLogger() Logger {
	return Logger{}
}

func (l *Logger) Info(msg string) {
	jsonHandler := slog.NewJSONHandler(os.Stderr, nil)
	myslog := slog.New(jsonHandler)
	myslog.Info(msg)
}

func (l *Logger) Error(msg string) {
	jsonHandler := slog.NewJSONHandler(os.Stderr, nil)
	myslog := slog.New(jsonHandler)
	myslog.Error(msg)
}
