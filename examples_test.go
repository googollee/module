package module_test

import (
	"context"
	"fmt"
	"regexp"

	"github.com/googollee/module"
)

type DB interface {
	Target() string
}

type db struct {
	target string
}

func (db *db) Target() string {
	return db.target
}

var (
	ModuleDB = module.New[DB]()
)

type Cache struct {
	fallback  DB
	keyPrefix string
}

var (
	ModuleCache  = module.New[*Cache]()
	ProvideCache = ModuleCache.ProvideWithFunc(func(ctx context.Context) (*Cache, error) {
		db := ModuleDB.Value(ctx)
		return &Cache{
			fallback:  db,
			keyPrefix: "cache",
		}, nil
	})
)

func ExampleModule() {
	repo := module.NewRepo()

	// No order required when adding providers
	repo.Add(ProvideCache)
	repo.Add(ModuleDB.ProvideValue(&db{target: "local.db"}))

	ctx := context.Background()

	ctx, err := repo.InjectTo(ctx)
	if err != nil {
		fmt.Println("inject error:", err)
		return
	}

	db := ModuleDB.Value(ctx)
	cache := ModuleCache.Value(ctx)

	fmt.Println("db target:", db.Target())
	fmt.Println("cache fallback target:", cache.fallback.Target())

	// Output:
	// db target: local.db
	// cache fallback target: local.db
}

func ExampleModule_loadOtherValue() {
	type Key string
	targetKey := Key("target")

	repo := module.NewRepo()
	repo.Add(ModuleDB.ProvideWithFunc(func(ctx context.Context) (DB, error) {
		// Load the target value from the context.
		target := ctx.Value(targetKey).(string)

		return &db{
			target: target,
		}, nil
	}))
	repo.Add(ProvideCache)

	// Store the target value in the context.
	ctx := context.WithValue(context.Background(), targetKey, "target.db")

	ctx, err := repo.InjectTo(ctx)
	if err != nil {
		fmt.Println("inject error:", err)
		return
	}

	db := ModuleDB.Value(ctx)
	cache := ModuleCache.Value(ctx)

	fmt.Println("db target:", db.Target())
	fmt.Println("cache fallback target:", cache.fallback.Target())

	// Output:
	// db target: target.db
	// cache fallback target: target.db
}

func ExampleModule_newPrefixInSpan() {
	repo := module.NewRepo()
	repo.Add(ModuleDB.ProvideValue(&db{target: "local.db"}))
	repo.Add(ProvideCache)

	ctx := context.Background()

	ctx, err := repo.InjectTo(ctx)
	if err != nil {
		fmt.Println("inject error:", err)
		return
	}

	db := ModuleDB.Value(ctx)
	cache := ModuleCache.Value(ctx)
	fmt.Println("before span, db target:", db.Target())
	fmt.Println("before span, cache prefix:", cache.keyPrefix)

	{
		// a new context in the span
		ctx := ModuleCache.With(ctx, &Cache{
			fallback:  db,
			keyPrefix: "span",
		})

		db := ModuleDB.Value(ctx)
		cache := ModuleCache.Value(ctx)
		fmt.Println("in span, db target:", db.Target())
		fmt.Println("in span, cache prefix:", cache.keyPrefix)
	}

	db = ModuleDB.Value(ctx)
	cache = ModuleCache.Value(ctx)
	fmt.Println("after span, db target:", db.Target())
	fmt.Println("after span, cache fallback target:", cache.keyPrefix)

	// Output:
	// before span, db target: local.db
	// before span, cache prefix: cache
	// in span, db target: local.db
	// in span, cache prefix: span
	// after span, db target: local.db
	// after span, cache fallback target: cache

}

func ExampleModule_createWithError() {
	repo := module.NewRepo()
	repo.Add(ModuleDB.ProvideValue(&db{target: "local.db"}))
	repo.Add(ModuleCache.ProvideWithFunc(func(ctx context.Context) (*Cache, error) {
		_ = ModuleDB.Value(ctx)
		return nil, fmt.Errorf("new cache error")
	}))

	ctx := context.Background()

	_, err := repo.InjectTo(ctx)
	if err != nil {
		fmt.Println("inject error:", err)
		return
	}

	// Output:
	// inject error: creating with module *module_test.Cache: new cache error
}

func ExampleModule_createWithPanic() {
	repo := module.NewRepo()
	repo.Add(ModuleDB.ProvideValue(&db{target: "localhost.db"}))
	repo.Add(ModuleCache.ProvideWithFunc(func(ctx context.Context) (*Cache, error) {
		_ = ModuleDB.Value(ctx)
		panic(fmt.Errorf("new cache error"))
	}))

	defer func() {
		err := recover()
		fmt.Println("panic:", err)
	}()

	ctx := context.Background()

	_, err := repo.InjectTo(ctx)
	if err != nil {
		fmt.Println("inject error:", err)
		return
	}

	// Output:
	// panic: new cache error
}

func ExampleModule_notExistingProvider() {
	ctx := context.Background()

	repo := module.NewRepo()
	repo.Add(ProvideCache)
	// repo.Add(ModuleDB.ProvideValue())

	_, err := repo.InjectTo(ctx)
	if err != nil {
		fmt.Println("inject error:", err)
		return
	}

	// Output:
	// inject error: creating with module module_test.DB: can't find module
}

func ExampleModule_duplicatingProviders() {
	defer func() {
		p := recover().(string)
		// Remove the file line info for testing.
		fmt.Println("panic:", regexp.MustCompile(`at .*`).ReplaceAllString(p, "at <removed file and line>"))
	}()

	repo := module.NewRepo()
	repo.Add(ModuleDB.ProvideValue(&db{target: "real.db"}))
	repo.Add(ModuleDB.ProvideValue(&db{target: "fake.db"}))

	// Output:
	// panic: already have a provider with type "module_test.DB", added at <removed file and line>
}
