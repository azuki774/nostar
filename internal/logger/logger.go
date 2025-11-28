package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime/debug"
)

// Default is the global logger used across the application.
// It behaves similarly to zap's SugaredLogger with key-value style.
var Default *slog.Logger

func init() {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}

	handler := slog.NewJSONHandler(os.Stderr, opts)
	Default = slog.New(handler)
}

// Info logs an informational message with structured key-value pairs.
func Info(msg string, args ...any) {
	Default.Info(msg, args...)
}

// Warn logs an informational message with structured key-value pairs.
func Warn(msg string, args ...any) {
	Default.Warn(msg, args...)
}

// Error logs an error message and attaches the error to the log record.
func Error(msg string, err error, args ...any) {
	if err != nil {
		args = append(args, "error", err)
	}
	Default.Error(msg, args...)
}

// Trace logs a very low-level message (below Debug) using slog.
// This gives you zap-like "trace" logging when you need it.
func Trace(msg string, args ...any) {
	Default.Log(context.Background(), slog.LevelDebug-4, msg, args...)
}

// LogPanic recovers from a panic (if any) and logs it with a stack trace,
// then exits with a non-zero status code.
func LogPanic() {
	if r := recover(); r != nil {
		Default.Error("panic recovered",
			"panic", r,
			"stack", string(debug.Stack()),
		)
		os.Exit(1)
	}
}
