package module

import (
	"context"
	"testing"
)

func TestWithError(t *testing.T) {
	wrapped := WithError(func(ctx context.Context) int {
		return 42
	})

	val, err := wrapped(context.Background())
	if err != nil {
		t.Errorf("WithError(f) returned error: %v", err)
	}
	if val != 42 {
		t.Errorf("WithError(f) returned %v, want 42", val)
	}
}

func TestCatchErrorPanic(t *testing.T) {
	repo := NewRepo()
	m := New[int]()

	repo.Add(m.ProvideWithFunc(func(ctx context.Context) (int, error) {
		panic("random panic")
	}))

	defer func() {
		r := recover()
		if r != "random panic" {
			t.Errorf("recover() = %v, want 'random panic'", r)
		}
	}()

	_, _ = repo.InjectTo(context.Background())
}
