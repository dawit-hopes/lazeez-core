package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint" 
)

// Logger interface defines our logging contract
type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)
}

type loggerImpl struct {
	svc *slog.Logger
}

// NewLogger initializes a logger.
// It uses "tint" for human-readable terminal output.
func NewLogger() Logger {
	var handler slog.Handler

	// In a real app, you'd check an ENV var here: if os.Getenv("ENV") == "production"
	// For now, let's set up the "Pretty" human-readable version:
	handler = tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.Kitchen, // "3:04PM" instead of long ISO dates
	})

	return &loggerImpl{
		svc: slog.New(handler),
	}
}

func (l *loggerImpl) Info(msg string, args ...any)  { l.svc.Info(msg, args...) }
func (l *loggerImpl) Error(msg string, args ...any) { l.svc.Error(msg, args...) }
func (l *loggerImpl) Debug(msg string, args ...any) { l.svc.Debug(msg, args...) }
func (l *loggerImpl) Warn(msg string, args ...any)  { l.svc.Warn(msg, args...) }
