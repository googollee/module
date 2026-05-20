package module_test

import (
	"context"
	"errors"
	"testing"

	"github.com/googollee/module"
)

type errorCloser struct{}

func (c *errorCloser) Close() error {
	return errors.New("cleanup error")
}

func TestRepo_Cleanup_Error(t *testing.T) {
	m := module.New[*errorCloser]()
	repo := module.NewRepo()
	repo.Add(m.ProvideValue(&errorCloser{}))

	ctx, err := repo.InjectTo(context.Background())
	if err != nil {
		t.Fatalf("failed to inject: %v", err)
	}

	// Trigger creation
	_ = m.Value(ctx)

	err = repo.Cleanup()
	if err == nil {
		t.Error("expected error during cleanup, got nil")
	}
	if err.Error() != "cleanup error" {
		t.Errorf("expected 'cleanup error', got '%v'", err)
	}
}

type manualCloser struct {
	closeFn func() error
}

func (c *manualCloser) Close() error {
	return c.closeFn()
}

func TestRepo_Cleanup_MultipleCalls(t *testing.T) {
	count := 0
	m := module.New[*manualCloser]()
	repo := module.NewRepo()
	repo.Add(m.ProvideValue(&manualCloser{
		closeFn: func() error {
			count++
			return nil
		},
	}))

	ctx, _ := repo.InjectTo(context.Background())
	_ = m.Value(ctx)

	_ = repo.Cleanup()
	_ = repo.Cleanup()

	if count != 1 {
		t.Errorf("expected Close to be called exactly once, called %d times", count)
	}
}
