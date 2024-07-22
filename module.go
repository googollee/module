// Package module provides a way to do dependency injection, with type-safe, without performance penalty.
// See [examples](#example-Module) for the basic usage.
package module

import (
	"context"
	"reflect"
)

type moduleKey string

// Module provides a module to inject and retreive an instance with its type.
type Module[T any] struct {
	moduleKey moduleKey
}

// New creates a new module with type `T` and the constructor `builder`.
func New[T any]() Module[T] {
	var t T
	return Module[T]{
		moduleKey: moduleKey(reflect.TypeOf(&t).Elem().String()),
	}
}

// Value returns an instance of T which is injected to the context.
func (m Module[T]) Value(ctx context.Context) T {
	var null T

	v := ctx.Value(m.moduleKey)
	if v == nil {
		return null
	}

	return v.(T)
}

func (m Module[T]) With(ctx context.Context, t T) context.Context {
	return context.WithValue(ctx, m.moduleKey, t)
}
