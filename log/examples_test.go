package log_test

import (
	"context"
	"log/slog"
	"os"

	"github.com/googollee/module"
	"github.com/googollee/module/log"
)

func removeTimeAttr(groups []string, a slog.Attr) slog.Attr {
	// Remove time from the output for predictable test output.
	if a.Key == slog.TimeKey {
		return slog.Attr{}
	}

	return a
}

func ExampleModule() {
	loggerOption := slog.HandlerOptions{
		AddSource:   false, // Remove code position from the output for predictable test output.
		Level:       slog.LevelDebug,
		ReplaceAttr: removeTimeAttr,
	}

	repo := module.NewRepo()
	// repo.Add(log.TextLogger)) in common usage
	// Provide a customed slog.Logger for predictable test output.
	repo.Add(log.Module.ProvideValue(slog.New(slog.NewTextHandler(os.Stdout, &loggerOption))))

	ctx, err := repo.InjectTo(context.Background())
	if err != nil {
		return
	}

	log.DEBUG(ctx, "debug")
	log.INFO(ctx, "info")
	log.WARN(ctx, "warning")
	log.ERROR(ctx, "error")

	// Output:
	// level=DEBUG msg=debug
	// level=INFO msg=info
	// level=WARN msg=warning
	// level=ERROR msg=error
}

func ExampleWith() {
	loggerOption := slog.HandlerOptions{
		AddSource:   false, // Remove code position from the output for predictable test output.
		ReplaceAttr: removeTimeAttr,
	}

	repo := module.NewRepo()
	// repo.Add(log.TextLogger)) in common usage
	// Provide a customed slog.Logger for predictable test output.
	repo.Add(log.Module.ProvideValue(slog.New(slog.NewTextHandler(os.Stdout, &loggerOption))))

	ctx, err := repo.InjectTo(context.Background())
	if err != nil {
		return
	}

	log.INFO(ctx, "before")
	{
		ctx := log.With(ctx, "span", "abc")
		log.INFO(ctx, "in")
	}
	log.INFO(ctx, "after")

	// Output:
	// level=INFO msg=before
	// level=INFO msg=in span=abc
	// level=INFO msg=after
}
