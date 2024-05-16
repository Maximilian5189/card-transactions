package logger

import (
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type Logger struct{}

func NewLogger() Logger {
	return Logger{}
}

func (l *Logger) Info(msg string) {
	_, file, line, _ := runtime.Caller(1)
	fileparts := strings.Split(file, "/")
	jsonHandler := slog.NewJSONHandler(os.Stderr, nil)
	myslog := slog.New(jsonHandler)
	myslog.Info(fileparts[len(fileparts)-1] + ":" + strconv.Itoa(line) + "_" + msg)
}

func (l *Logger) Error(msg string) {
	_, file, line, _ := runtime.Caller(1)
	fileparts := strings.Split(file, "/")
	jsonHandler := slog.NewJSONHandler(os.Stderr, nil)
	myslog := slog.New(jsonHandler)
	myslog.Error(fileparts[len(fileparts)-1] + ":" + strconv.Itoa(line) + "_" + msg)
}
