package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	instance *slog.Logger
	once     sync.Once
)

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

func toSlogLevel(level Level) slog.Level {
	switch level {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Init(level Level) {
	once.Do(func() {
		opts := &slog.HandlerOptions{
			Level: toSlogLevel(level),
		}

		handler := slog.NewJSONHandler(os.Stdout, opts)
		instance = slog.New(handler)
		slog.SetDefault(instance)
	})
}

func Get() *slog.Logger {
	if instance == nil {
		Init(LevelInfo)
	}
	return instance
}

func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

func With(args ...any) *slog.Logger {
	return Get().With(args...)
}
