package groute

import (
	"net/http"
	"strings"
)

// Router represents a route router with shared middleware and prefix.
type Router struct {
	prefix      string
	middlewares []Middleware
	mux         *http.ServeMux
}

// NewRouter creates a new router.
func NewRouter() *Router {
	return &Router{
		mux:         http.NewServeMux(),
		middlewares: make([]Middleware, 0),
	}
}

// Use adds middleware to the router.
// Middleware will be applied in the order they are added.
func (g *Router) Use(middlewares ...Middleware) {
	g.middlewares = append(g.middlewares, middlewares...)
}

// Get registers a GET route.
func (g *Router) Get(pattern string, handler http.HandlerFunc) {
	g.HandleFunc("GET "+pattern, handler)
}

// Post registers a POST route.
func (g *Router) Post(pattern string, handler http.HandlerFunc) {
	g.HandleFunc("POST "+pattern, handler)
}

// Put registers a PUT route.
func (g *Router) Put(pattern string, handler http.HandlerFunc) {
	g.HandleFunc("PUT "+pattern, handler)
}

// Delete registers a DELETE route.
func (g *Router) Delete(pattern string, handler http.HandlerFunc) {
	g.HandleFunc("DELETE "+pattern, handler)
}

// Patch registers a PATCH route.
func (g *Router) Patch(pattern string, handler http.HandlerFunc) {
	g.HandleFunc("PATCH "+pattern, handler)
}

// Head registers a HEAD route.
func (g *Router) Head(pattern string, handler http.HandlerFunc) {
	g.HandleFunc("HEAD "+pattern, handler)
}

// Options registers an OPTIONS route.
func (g *Router) Options(pattern string, handler http.HandlerFunc) {
	g.HandleFunc("OPTIONS "+pattern, handler)
}

// Connect registers a CONNECT route.
func (g *Router) Connect(pattern string, handler http.HandlerFunc) {
	g.HandleFunc("CONNECT "+pattern, handler)
}

// Trace registers a TRACE route.
func (g *Router) Trace(pattern string, handler http.HandlerFunc) {
	g.HandleFunc("TRACE "+pattern, handler)
}

// Handle registers a route with any HTTP method.
func (g *Router) Handle(pattern string, handler http.Handler) {
	fullPattern := joinPath(g.prefix, pattern)
	// Apply middlewares to handler
	wrappedHandler := g.applyMiddlewares(handler)
	g.mux.Handle(fullPattern, wrappedHandler)
}

// HandleFunc registers a route handler function.
func (g *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	g.Handle(pattern, http.HandlerFunc(handler))
}

// ServeHTTP implements http.Handler interface.
func (g *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

// Group creates a sub-group with additional prefix and middleware.
func (g *Router) Group(prefix string) *Router {
	subGroupPrefix := strings.TrimRight(g.prefix, "/") + "/" + strings.TrimLeft(prefix, "/")

	subGroup := &Router{
		prefix:      subGroupPrefix,
		mux:         g.mux,
		middlewares: make([]Middleware, len(g.middlewares)),
	}
	// Copy parent middlewares
	copy(subGroup.middlewares, g.middlewares)

	return subGroup
}

// applyMiddlewares applies all middlewares to a handler.
func (g *Router) applyMiddlewares(handler http.Handler) http.Handler {
	// Apply middlewares in reverse order (first added = outermost)
	// This ensures the first middleware added executes first.
	h := http.HandlerFunc(handler.ServeHTTP)
	for i := len(g.middlewares) - 1; i >= 0; i-- {
		h = g.middlewares[i](h)
	}
	return h
}

// joinPath joins prefix and pattern, ensuring proper slash handling.
// Pattern may contain HTTP method prefix like "GET /path" or just "/path".
func joinPath(prefix, pattern string) string {
	// Handle pattern with HTTP method prefix (e.g., "GET /users")
	parts := strings.SplitN(pattern, " ", 2)
	methodPrefix := ""
	path := pattern
	if len(parts) == 2 {
		methodPrefix = parts[0] + " "
		path = parts[1]
	}

	path = strings.TrimRight(prefix, "/") + "/" + strings.TrimLeft(path, "/")

	return methodPrefix + path
}
