package module

import "context"

// WithError wraps a function with returning an always-nil error.
func WithError[T any](f func(context.Context) T) func(context.Context) (T, error) {
	return func(ctx context.Context) (T, error) {
		return f(ctx), nil
	}
}

// Provider is the interface to provide an Instance.
type Provider interface {
	key() moduleKey
	value(ctx context.Context) (any, error)
}

// BuildFunc is the constructor of an Instance.
type BuildFunc[T any] func(context.Context) (T, error)

type funcProvider[T any] struct {
	moduleKey moduleKey
	ctor      BuildFunc[T]
}

// ProvideWithFunc returns a provider which provides instances creating from `ctor` function.
func (m Module[T]) ProvideWithFunc(ctor BuildFunc[T]) Provider {
	return &funcProvider[T]{
		moduleKey: m.moduleKey,
		ctor:      ctor,
	}
}

// ProvideValue returns a provider which always provides given `value` as instances.
func (m Module[T]) ProvideValue(value T) Provider {
	return m.ProvideWithFunc(func(context.Context) (T, error) {
		return value, nil
	})
}

func (p funcProvider[T]) key() moduleKey {
	return p.moduleKey
}

func (p funcProvider[T]) value(ctx context.Context) (any, error) {
	return p.ctor(ctx)
}
