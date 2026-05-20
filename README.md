[![Go Reference](https://pkg.go.dev/badge/github.com/googollee/module.svg)](https://pkg.go.dev/github.com/googollee/module)
[![License Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fgoogollee%2Fmodule.svg?type=shield&issueType=license)](https://app.fossa.com/projects/git%2Bgithub.com%2Fgoogollee%2Fmodule?ref=badge_shield&issueType=license)
[![Security Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fgoogollee%2Fmodule.svg?type=shield&issueType=security)](https://app.fossa.com/projects/git%2Bgithub.com%2Fgoogollee%2Fmodule?ref=badge_shield&issueType=security)
[![Tests](https://github.com/googollee/module/actions/workflows/tests.yml/badge.svg)](https://github.com/googollee/module/actions/workflows/tests.yml)
[![codecov](https://codecov.io/gh/googollee/module/graph/badge.svg?token=M379D3EK0I)](https://codecov.io/gh/googollee/module)

# Module

A type-safe, performance-focused dependency injection framework for Go, built on top of `context.Context`.

## Key Features

- **Type-Safe**: No `interface{}` type assertions in your business logic.
- **High Performance**: No reflection penalty at runtime; retrieval is a simple map lookup.
- **Context-Based**: Seamlessly integrates with standard Go middleware and request flows.
- **Robust**: Built-in circular dependency detection and deterministic injection order.
- **Reliable**: 100% test coverage.

## Quick Start

```go
// 1. Define a module token
var ModuleDB = module.New[DB]()

// 2. Register a provider
repo := module.NewRepo()
repo.Add(ModuleDB.ProvideValue(&db{target: "local.db"}))

// 3. Inject and use
ctx, _ := repo.InjectTo(context.Background())
database := ModuleDB.Value(ctx) // Returns DB type directly
```

For more detailed usage, including dependency chains and scoped overrides, see the [Examples](./examples_test.go) or [![Go Reference](https://pkg.go.dev/badge/github.com/googollee/module.svg)](https://pkg.go.dev/github.com/googollee/module).

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fgoogollee%2Fmodule.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fgoogollee%2Fmodule?ref=badge_large)
