# GRoute

[English](README.md) | 中文

轻量级 Go 路由封装，基于标准库 `net/http` 与 `http.ServeMux`（支持 Go 1.22+ 的路径参数与通配符）。

## 特性

- **零依赖**：仅使用 Go 标准库
- **`http.Handler` 兼容**：`Router` 实现 `ServeHTTP`
- **中间件**：支持链式中间件（`func(http.HandlerFunc) http.HandlerFunc`）
- **分组**：支持前缀分组，子组会继承父组中间件
- **路径参数与通配符**：基于 `http.ServeMux`（例如 `/user/{id}`、`/{path...}`）

## 安装

```bash
go get github.com/lyuangg/grouter
```

## 快速开始

```go
package main

import (
	"net/http"

	"github.com/lyuangg/grouter"
)

func main() {
	r := groute.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello, World!"))
	})

	_ = http.ListenAndServe(":8080", r)
}
```

## 注册路由

你可以用便捷方法注册特定 HTTP 方法，也可以用 `Handle/HandleFunc` 注册任意模式。

```go
r := groute.NewRouter()

r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

// 也可以显式带上方法前缀
r.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
})
```

## 路径参数

GRoute 直接使用标准库 `http.ServeMux` 的路径参数；在 handler 里用 `r.PathValue(name)` 取值。

```go
r := groute.NewRouter()

r.Get("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, _ = w.Write([]byte(id))
})
```

## 通配符

```go
r := groute.NewRouter()

r.Get("/{pathname...}", func(w http.ResponseWriter, r *http.Request) {
	p := r.PathValue("pathname")
	_, _ = w.Write([]byte(p))
})
```

## 路由分组

`Group(prefix)` 会创建一个共享同一个底层 mux 的子路由器，并自动拼接前缀；子组会继承父组中间件。

```go
r := groute.NewRouter()

api := r.Group("/api")
api.Get("/users", func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

v1 := api.Group("/v1")
v1.Get("/info", func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})
```

## 中间件

中间件类型：

```go
type Middleware func(http.HandlerFunc) http.HandlerFunc
```

使用示例：

```go
r := groute.NewRouter()

r.Use(func(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Powered-By", "groute")
		next(w, r)
	}
})

r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})
```

## 兼容性说明

- **Go 版本**：路径参数/通配符依赖 Go 1.22+ 的 `http.ServeMux` 行为。
- **路由匹配**：实际匹配规则由标准库 `http.ServeMux` 决定。

## License

MIT
