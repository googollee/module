package mlog

import (
	"context"
	"log/slog"
	"os"

	"github.com/googollee/module"
)

var (
	Module = module.New[*slog.Logger]()

	TextLogger = Module.ProvideWithFunc(func(ctx context.Context) (*slog.Logger, error) {
		handler := slog.NewTextHandler(os.Stderr, nil)
		return slog.New(handler), nil
	})
	JSONLogger = Module.ProvideWithFunc(func(ctx context.Context) (*slog.Logger, error) {
		handler := slog.NewJSONHandler(os.Stderr, nil)
		return slog.New(handler), nil
	})
)

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

func DEBUG(ctx context.Context, msg string, args ...any) {
	logger := Module.Value(ctx)
	if logger == nil {
		return
	}

	logger.DebugContext(ctx, msg, args...)
}

func INFO(ctx context.Context, msg string, args ...any) {
	logger := Module.Value(ctx)
	if logger == nil {
		return
	}

	logger.InfoContext(ctx, msg, args...)
}

func WARN(ctx context.Context, msg string, args ...any) {
	logger := Module.Value(ctx)
	if logger == nil {
		return
	}

	logger.WarnContext(ctx, msg, args...)
}

func ERROR(ctx context.Context, msg string, args ...any) {
	logger := Module.Value(ctx)
	if logger == nil {
		return
	}

	logger.ErrorContext(ctx, msg, args...)
}
