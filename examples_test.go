package module_test

import (
	"context"
	"fmt"
	"regexp"

	"github.com/googollee/module"
)

// --- Basic Types for Examples ---

type DB interface {
	Target() string
}

type db struct {
	target string
}

func (db *db) Target() string {
	return db.target
}

type CloserDB struct {
	target string
}

func (c *CloserDB) Target() string {
	return c.target
}

func (c *CloserDB) Close() error {
	fmt.Printf("closing db: %s\n", c.target)
	return nil
}

type Cache struct {
	fallback  DB
	keyPrefix string
}

// --- Module Definitions (Global Tokens) ---

var (
	// ModuleDB is a token for DB interface.
	ModuleDB = module.New[DB]()

	// ModuleCache is a token for *Cache struct.
	ModuleCache = module.New[*Cache]()

	// ProvideCache defines how to create a Cache.
	// It depends on ModuleDB.
	ProvideCache = ModuleCache.ProvideWithFunc(func(ctx context.Context) (*Cache, error) {
		// Use Module.Value to get dependencies from the context during construction.
		db := ModuleDB.Value(ctx)
		return &Cache{
			fallback:  db,
			keyPrefix: "cache",
		}, nil
	})
)

// ExampleModule_basic shows the simplest way to use the library:
// Define a module, add its provider, inject it, and use it.
func ExampleModule_basic() {
	// 1. Create a repository to hold providers.
	repo := module.NewRepo()

	// 2. Add a provider (in this case, a static value).
	repo.Add(ModuleDB.ProvideValue(&db{target: "simple.db"}))

	// 3. Inject providers into a context.
	// This creates the instances defined by the providers.
	ctx, err := repo.InjectTo(context.Background())
	if err != nil {
		fmt.Println("inject error:", err)
		return
	}

	// 4. Retrieve the instance using the Module token.
	// No type assertion needed!
	database := ModuleDB.Value(ctx)
	fmt.Println("db target:", database.Target())

	// Output:
	// db target: simple.db
}

// ExampleModule_dependencies demonstrates automatic dependency resolution.
// Even if ProvideCache is added before ModuleDB, the library handles it.
func ExampleModule_dependencies() {
	repo := module.NewRepo()

	// No order required when adding providers.
	// ModuleCache depends on ModuleDB (see ProvideCache definition above).
	repo.Add(ProvideCache)
	repo.Add(ModuleDB.ProvideValue(&db{target: "local.db"}))

	ctx, err := repo.InjectTo(context.Background())
	if err != nil {
		fmt.Println("inject error:", err)
		return
	}

	// Retrieve the cache. Its 'fallback' field will be automatically populated with the DB.
	cache := ModuleCache.Value(ctx)
	fmt.Println("cache fallback target:", cache.fallback.Target())

	// Output:
	// cache fallback target: local.db
}

// ExampleModule_contextValues shows how a provider can access existing values in the context.
func ExampleModule_contextValues() {
	type Key string
	targetKey := Key("target")

	repo := module.NewRepo()
	repo.Add(ModuleDB.ProvideWithFunc(func(ctx context.Context) (DB, error) {
		// Providers can read standard context values.
		target := ctx.Value(targetKey).(string)
		return &db{target: target}, nil
	}))

	// Put a value into the context before injection.
	ctx := context.WithValue(context.Background(), targetKey, "from_ctx.db")

	ctx, err := repo.InjectTo(ctx)
	if err != nil {
		fmt.Println("inject error:", err)
		return
	}

	database := ModuleDB.Value(ctx)
	fmt.Println("db target:", database.Target())

	// Output:
	// db target: from_ctx.db
}

// ExampleModule_scopedOverride demonstrates how to override a module instance
// for a specific scope using Module.With.
func ExampleModule_scopedOverride() {
	repo := module.NewRepo()
	repo.Add(ModuleDB.ProvideValue(&db{target: "default.db"}))
	repo.Add(ProvideCache)

	ctx, _ := repo.InjectTo(context.Background())

	fmt.Println("original prefix:", ModuleCache.Value(ctx).keyPrefix)

	{
		// Create a new context where ModuleCache is overridden.
		scopedCtx := ModuleCache.With(ctx, &Cache{
			fallback:  ModuleDB.Value(ctx),
			keyPrefix: "scoped",
		})

		fmt.Println("scoped prefix:", ModuleCache.Value(scopedCtx).keyPrefix)
	}

	// The original context remains unchanged.
	fmt.Println("back to original prefix:", ModuleCache.Value(ctx).keyPrefix)

	// Output:
	// original prefix: cache
	// scoped prefix: scoped
	// back to original prefix: cache
}

// --- Error Handling Examples ---

func ExampleModule_errorHandling() {
	repo := module.NewRepo()

	// Provider returns an error.
	repo.Add(ModuleDB.ProvideWithFunc(func(ctx context.Context) (DB, error) {
		return nil, fmt.Errorf("connection failed")
	}))

	_, err := repo.InjectTo(context.Background())
	if err != nil {
		fmt.Println("inject error:", err)
	}

	// Output:
	// inject error: creating with module module_test.DB: connection failed
}

func ExampleModule_circularDependency() {
	type A struct{}
	type B struct{}

	mA := module.New[*A]()
	mB := module.New[*B]()

	repo := module.NewRepo()
	repo.Add(mA.ProvideWithFunc(func(ctx context.Context) (*A, error) {
		_ = mB.Value(ctx)
		return &A{}, nil
	}))
	repo.Add(mB.ProvideWithFunc(func(ctx context.Context) (*B, error) {
		_ = mA.Value(ctx)
		return &B{}, nil
	}))

	_, err := repo.InjectTo(context.Background())
	if err != nil {
		fmt.Println("inject error:", err)
	}

	// Output:
	// inject error: creating with module *module_test.A: circular dependency detected: *module_test.A -> *module_test.B -> *module_test.A
}

func ExampleModule_registrationErrors() {
	// 1. Missing provider
	repo1 := module.NewRepo()
	repo1.Add(ProvideCache) // Depends on DB, but DB provider is missing.
	_, err := repo1.InjectTo(context.Background())
	fmt.Println("missing provider:", err)

	// 2. Duplicate provider (causes panic on Add)
	repo2 := module.NewRepo()
	repo2.Add(ModuleDB.ProvideValue(&db{target: "1"}))

	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("%v", r)
			fmt.Println("duplicate provider:", regexp.MustCompile(`at .*`).ReplaceAllString(msg, "at <removed>"))
		}
	}()
	repo2.Add(ModuleDB.ProvideValue(&db{target: "2"}))

	// Output:
	// missing provider: creating with module module_test.DB: can't find module
	// duplicate provider: already have a provider with type module_test.DB, added at <removed>
}
// ExampleModule_cleanup demonstrates the automatic cleanup feature.
// If an injected instance implements `interface{ Close() error }`,
// repo.Cleanup() will automatically call it in reverse order of creation.
func ExampleModule_cleanup() {
	repo := module.NewRepo()
	defer func() {
		if err := repo.Cleanup(); err != nil {
			fmt.Println("cleanup error:", err)
		}
	}()

	// 1. Add a provider that returns a 'closer' (CloserDB implements Close()).
	repo.Add(ModuleDB.ProvideValue(&CloserDB{target: "closable.db"}))

	// 2. Add a provider that returns a non-closer (Cache does not have Close()).
	repo.Add(ProvideCache)

	// 3. Inject and use.
	ctx, _ := repo.InjectTo(context.Background())

	// Trigger creation of both modules.
	// Order of creation: CloserDB (dependency), then Cache.
	_ = ModuleCache.Value(ctx)

	fmt.Println("doing work...")

	// Output:
	// doing work...
	// closing db: closable.db
}
