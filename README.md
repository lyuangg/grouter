# GRouter

English | [中文](README.zh-CN.md)

A lightweight router wrapper built on Go standard library `net/http` and `http.ServeMux` (supports Go 1.22+ path parameters and wildcards).

## Features

- **Zero dependencies**: standard library only
- **`http.Handler` compatible**: `Router` implements `ServeHTTP`
- **Middleware**: chainable middlewares (`func(http.HandlerFunc) http.HandlerFunc`)
- **Grouping**: prefix-based groups; sub-groups inherit middlewares
- **Path params & wildcards**: powered by `http.ServeMux` (e.g. `/user/{id}`, `/{path...}`)

## Install

```bash
go get github.com/lyuangg/grouter
```

## Quick start

```go
package main

import (
	"net/http"

	"github.com/lyuangg/grouter"
)

func main() {
	r := grouter.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello, World!"))
	})

	_ = http.ListenAndServe(":8080", r)
}
```

## Register routes

You can register method-specific routes via helpers, or use `Handle/HandleFunc` for arbitrary patterns.

```go
r := grouter.NewRouter()

r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

// You can also include the method prefix explicitly.
r.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
})
```

## Path parameters

GRoute uses Go standard library `http.ServeMux` path params; use `r.PathValue(name)` in handlers.

```go
r := grouter.NewRouter()

r.Get("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, _ = w.Write([]byte(id))
})
```

## Wildcards

```go
r := grouter.NewRouter()

r.Get("/{pathname...}", func(w http.ResponseWriter, r *http.Request) {
	p := r.PathValue("pathname")
	_, _ = w.Write([]byte(p))
})
```

## Route grouping

`Group(prefix)` creates a sub-router sharing the same underlying mux, with an added path prefix; middlewares are inherited.

```go
r := grouter.NewRouter()

api := r.Group("/api")
api.Get("/users", func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

v1 := api.Group("/v1")
v1.Get("/info", func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})
```

## Middleware

Middleware type:

```go
type Middleware func(http.HandlerFunc) http.HandlerFunc
```

Example:

```go
r := grouter.NewRouter()

r.Use(func(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Powered-By", "grouter")
		next(w, r)
	}
})

r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})
```

## Compatibility notes

- **Go version**: path params/wildcards require Go 1.22+ `http.ServeMux` behavior.
- **Routing behavior**: matching rules are defined by the standard library `http.ServeMux`.

## License

MIT
