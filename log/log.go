// Package log is a module injecting a `*slog.Logger` instance.
// It provides a group of functions to log with the injected instance.
// See [examples_test.go](./examples_test.go) for the usage.
package log

import (
	"context"
	"log/slog"
	"os"

	"github.com/googollee/module"
)

var (
	// Module for injecting `*slog.Logger`
	Module = module.New[*slog.Logger]()

	// TextLogger provides a instance with `slog.TextHandler` logging to `os.Stderr`
	TextLogger = Module.ProvideWithFunc(func(ctx context.Context) (*slog.Logger, error) {
		handler := slog.NewTextHandler(os.Stderr, nil)
		return slog.New(handler), nil
	})

	// JSONLogger provides a instance with `slog.JSONHandler` logging to `os.Stderr`
	JSONLogger = Module.ProvideWithFunc(func(ctx context.Context) (*slog.Logger, error) {
		handler := slog.NewJSONHandler(os.Stderr, nil)
		return slog.New(handler), nil
	})
)

// With creates a new `context.Context` with new attrs. It's similar to [`slog.Logger.With()`](https://pkg.go.dev/log/slog#Logger.With).
func With(ctx context.Context, args ...any) context.Context {
	if len(args) == 0 {
		return ctx
	}

	logger := Module.Value(ctx)
	if logger == nil {
		return ctx
	}

	logger = logger.With(args...)

	return Module.With(ctx, logger)
}

// DEBUG logs with the injected instance at DEBUG level.
// If no injected `*slog.Logger`, the function does nothing.
func DEBUG(ctx context.Context, msg string, args ...any) {
	logger := Module.Value(ctx)
	if logger == nil {
		return
	}

	logger.DebugContext(ctx, msg, args...)
}

// INFO logs with the injected instance at INFO level.
// If no injected `*slog.Logger`, the function does nothing.
func INFO(ctx context.Context, msg string, args ...any) {
	logger := Module.Value(ctx)
	if logger == nil {
		return
	}

	logger.InfoContext(ctx, msg, args...)
}

// WARN logs with the injected instance at WARN level.
// If no injected `*slog.Logger`, the function does nothing.
func WARN(ctx context.Context, msg string, args ...any) {
	logger := Module.Value(ctx)
	if logger == nil {
		return
	}

	logger.WarnContext(ctx, msg, args...)
}

// ERROR logs with the injected instance at ERROR level.
// If no injected `*slog.Logger`, the function does nothing.
func ERROR(ctx context.Context, msg string, args ...any) {
	logger := Module.Value(ctx)
	if logger == nil {
		return
	}

	logger.ErrorContext(ctx, msg, args...)
}
