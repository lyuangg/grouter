package groute

import (
	"net/http"
)

// Middleware wraps a handler function.
//
// Using http.HandlerFunc makes it convenient to call next as next(w, r).
type Middleware func(http.HandlerFunc) http.HandlerFunc
