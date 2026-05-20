package module

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"runtime"
	"slices"
	"sync"
)

type providerWithLine struct {
	provider Provider
	file     string
	line     int
}

// Repo is a repository of modules, and to inject instances creating by modules into a context.
type Repo struct {
	providers map[moduleKey]providerWithLine

	init      sync.Once
	instances map[moduleKey]any

	cleanups    []func() error
	cleanupOnce sync.Once
}

// NewRepo creates a Repo instance.
func NewRepo() *Repo {
	return &Repo{
		providers: make(map[moduleKey]providerWithLine),
		instances: make(map[moduleKey]any),
	}
}

// Add adds a provider to the repo.
func (r *Repo) Add(provider Provider) {
	if p, ok := r.providers[provider.key()]; ok {
		msg := fmt.Sprintf("already have a provider with type %s, added at %s:%d", provider.key(), p.file, p.line)
		panic(msg)
	}

	_, file, line, _ := runtime.Caller(1)
	r.providers[provider.key()] = providerWithLine{
		provider: provider,
		file:     file,
		line:     line,
	}
}

func (r *Repo) addCleanup(cleanup func() error) {
	r.cleanups = append(r.cleanups, cleanup)
}

// InjectTo injects instances created by modules into a context `ctx`.
// It returns a new context with all injections. If any module creates an instance with an error, `InjectTo` returns that error with the module name.
// InjectTo is not thread-safe. It should not be called concurrently.
// InjectTo should not be called concurrently with Cleanup.
// Injecting instances only create once if necessary. Calling `InjectTo` mutlple times share instances between returning contexts.
// InjectTo ignores all new providers adding to the Repo after the first run. So adding all providers before calling `InjectTo`.
func (r *Repo) InjectTo(ctx context.Context) (context.Context, error) {
	var err error
	r.init.Do(func() {
		err = r.buildValues(ctx)
	})
	if err != nil {
		return nil, err
	}

	return &moduleContext{
		Context:   ctx,
		instances: r.instances,
	}, nil
}

// Cleanup executes all Close methods of injected instances in reverse order of creation.
// Cleanup should not be called concurrently with InjectTo.
// Cleanup only runs once. Subsequent calls do nothing and return nil.
func (r *Repo) Cleanup() (err error) {
	r.cleanupOnce.Do(func() {
		var errs []error
		for i := len(r.cleanups) - 1; i >= 0; i-- {
			if e := r.cleanups[i](); e != nil {
				errs = append(errs, e)
			}
		}
		r.cleanups = nil
		err = errors.Join(errs...)
	})

	return
}

func (r *Repo) buildValues(ctx context.Context) (err error) {
	defer func() {
		err = r.catchError(recover())
	}()

	builder := &buildContext{
		repo:      r,
		Context:   ctx,
		providers: r.providers,
		instances: r.instances,
	}

	keys := make([]moduleKey, 0, len(r.providers))
	for key := range r.providers {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(a, b moduleKey) int {
		return cmp.Compare(a.String(), b.String())
	})

	for _, key := range keys {
		_ = builder.Value(key)
	}

	return
}

func (r *Repo) catchError(err any) error {
	if err == nil {
		return nil
	}

	createErr, ok := err.(createPanic)
	if !ok {
		panic(err)
	}

	return fmt.Errorf("creating with module %s: %w", createErr.key, createErr.err)
}
