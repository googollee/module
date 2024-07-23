package log_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"regexp"
	"testing"

	"github.com/googollee/module"
	"github.com/googollee/module/log"
)

func captureStderr(t *testing.T, f func(t *testing.T)) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal("create pipe error:", err)
	}

	orig := os.Stderr
	os.Stderr = w

	defer func() {
		os.Stderr = orig
		w.Close()
	}()

	f(t)

	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal("read pipe error:", err)
	}

	return string(out)
}

func TestTextHandler(t *testing.T) {
	output := captureStderr(t, func(t *testing.T) {
		repo := module.NewRepo()
		repo.Add(log.TextLogger)

		ctx, err := repo.InjectTo(context.Background())
		if err != nil {
			t.Fatal("repo.InjectTo() error:", err)
		}

		log.ERROR(ctx, "log")
	})

	output = regexp.MustCompile("time=[^ ]*").ReplaceAllString(output, "time=<time>")

	if got, want := output, "time=<time> level=ERROR msg=log\n"; got != want {
		t.Errorf("\ngot : %q\nwant: %q", got, want)
	}
}

func TestJSONHandler(t *testing.T) {
	output := captureStderr(t, func(t *testing.T) {
		repo := module.NewRepo()
		repo.Add(log.JSONLogger)

		ctx, err := repo.InjectTo(context.Background())
		if err != nil {
			t.Fatal("repo.InjectTo() error:", err)
		}

		log.ERROR(ctx, "log")
	})

	output = regexp.MustCompile(`"time":"[^"]*"`).ReplaceAllString(output, `"time":"<time>"`)

	if got, want := output, "{\"time\":\"<time>\",\"level\":\"ERROR\",\"msg\":\"log\"}\n"; got != want {
		t.Errorf("\ngot : %q\nwant: %q", got, want)
	}
}

func TestNoInjection(t *testing.T) {
	output := captureStderr(t, func(t *testing.T) {
		ctx := context.Background()
		log.DEBUG(ctx, "debug")
		log.INFO(ctx, "info")
		log.WARN(ctx, "warning")
		log.ERROR(ctx, "error")

		ctx = log.With(ctx, "span", "in_span")

		log.DEBUG(ctx, "debug")
		log.INFO(ctx, "info")
		log.WARN(ctx, "warning")
		log.ERROR(ctx, "error")
	})

	if got, want := output, ""; got != want {
		t.Errorf("got: %q, want: %q", got, want)
	}
}

func TestWithNoAttr(t *testing.T) {
	var buf bytes.Buffer

	loggerOption := slog.HandlerOptions{
		AddSource:   false, // Remove code position from the output for predictable test output.
		ReplaceAttr: removeTimeAttr,
	}

	repo := module.NewRepo()
	repo.Add(log.Module.ProvideValue(slog.New(slog.NewTextHandler(&buf, &loggerOption))))

	ctx, err := repo.InjectTo(context.Background())
	if err != nil {
		return
	}

	log.INFO(ctx, "before")
	{
		ctx := log.With(ctx)
		log.INFO(ctx, "in")
	}
	log.INFO(ctx, "after")

	if got, want := buf.String(), "level=INFO msg=before\nlevel=INFO msg=in\nlevel=INFO msg=after\n"; got != want {
		t.Errorf("\ngot: %q\nwant: %q", got, want)
	}
}
