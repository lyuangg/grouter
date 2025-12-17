package groute

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRouter(t *testing.T) {
	g := NewRouter()
	if g == nil {
		t.Fatal("NewRouter() returned nil")
	}
	if g.mux == nil {
		t.Error("mux should not be nil")
	}
	if g.middlewares == nil {
		t.Error("middlewares should not be nil")
	}
	if len(g.middlewares) != 0 {
		t.Errorf("expected 0 middlewares, got %d", len(g.middlewares))
	}
	if g.prefix != "" {
		t.Errorf("expected empty prefix, got %q", g.prefix)
	}
}

func TestHTTPMethods(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		registerMethod func(*Router, string, http.HandlerFunc)
		expectedStatus int
	}{
		{"GET", "GET", (*Router).Get, http.StatusOK},
		{"POST", "POST", (*Router).Post, http.StatusCreated},
		{"PUT", "PUT", (*Router).Put, http.StatusOK},
		{"DELETE", "DELETE", (*Router).Delete, http.StatusOK},
		{"PATCH", "PATCH", (*Router).Patch, http.StatusOK},
		{"HEAD", "HEAD", (*Router).Head, http.StatusOK},
		{"OPTIONS", "OPTIONS", (*Router).Options, http.StatusOK},
		{"CONNECT", "CONNECT", (*Router).Connect, http.StatusOK},
		{"TRACE", "TRACE", (*Router).Trace, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewRouter()
			called := false
			tt.registerMethod(g, "/test", func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(tt.expectedStatus)
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)

			if !called {
				t.Error("handler was not called")
			}
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestMethodMismatch(t *testing.T) {
	g := NewRouter()
	called := false
	g.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	// Try POST request to GET route
	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)

	// Handler should not be called for wrong method
	if called {
		t.Error("handler should not be called for wrong HTTP method")
	}
}

func TestHandle(t *testing.T) {
	g := NewRouter()
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	g.Handle("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)

	if !called {
		t.Error("handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleFunc(t *testing.T) {
	g := NewRouter()
	called := false
	g.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)

	if !called {
		t.Error("handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestUse(t *testing.T) {
	g := NewRouter()
	order := []string{}

	middleware1 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "middleware1")
			next(w, r)
		}
	}

	middleware2 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "middleware2")
			next(w, r)
		}
	}

	g.Use(middleware1, middleware2)
	g.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)

	// First added middleware should execute first
	expectedOrder := []string{"middleware1", "middleware2", "handler"}
	if len(order) != len(expectedOrder) {
		t.Errorf("expected %d calls, got %d: %v", len(expectedOrder), len(order), order)
	}
	for i, expected := range expectedOrder {
		if i < len(order) && order[i] != expected {
			t.Errorf("expected order[%d] = %q, got %q", i, expected, order[i])
		}
	}
}

func TestMultipleRoutes(t *testing.T) {
	g := NewRouter()
	getCalled := false
	postCalled := false

	g.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		getCalled = true
		w.WriteHeader(http.StatusOK)
	})

	g.Post("/test", func(w http.ResponseWriter, r *http.Request) {
		postCalled = true
		w.WriteHeader(http.StatusCreated)
	})

	// Test GET
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)
	if !getCalled {
		t.Error("GET handler was not called")
	}

	// Test POST
	req = httptest.NewRequest("POST", "/test", nil)
	w = httptest.NewRecorder()
	g.ServeHTTP(w, req)
	if !postCalled {
		t.Error("POST handler was not called")
	}
}

func TestGroup(t *testing.T) {
	tests := []struct {
		name        string
		prefix      string
		route       string
		requestPath string
	}{
		{
			name:        "basic group",
			prefix:      "/api",
			route:       "/users",
			requestPath: "/api/users",
		},
		{
			name:        "prefix with trailing slash",
			prefix:      "/api/",
			route:       "/users",
			requestPath: "/api/users",
		},
		{
			name:        "empty prefix",
			prefix:      "",
			route:       "/test",
			requestPath: "/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewRouter()
			called := false

			subGroup := g.Group(tt.prefix)
			subGroup.Get(tt.route, func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", tt.requestPath, nil)
			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)

			if !called {
				t.Error("handler was not called")
			}
			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}
		})
	}
}

func TestGroupNested(t *testing.T) {
	g := NewRouter()
	called := false

	apiGroup := g.Group("/api")
	v1Group := apiGroup.Group("/v1")
	v1Group.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)

	if !called {
		t.Error("handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGroupInheritsMiddlewares(t *testing.T) {
	g := NewRouter()
	order := []string{}

	parentMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "parent")
			next(w, r)
		}
	}

	childMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "child")
			next(w, r)
		}
	}

	g.Use(parentMiddleware)
	subGroup := g.Group("/api")
	subGroup.Use(childMiddleware)
	subGroup.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)

	// Middleware execution order: first added = outermost (executes first)
	// So parent middleware (added first) executes before child middleware
	expectedOrder := []string{"parent", "child", "handler"}
	if len(order) != len(expectedOrder) {
		t.Errorf("expected %d calls, got %d: %v", len(expectedOrder), len(order), order)
	}
	for i, expected := range expectedOrder {
		if i < len(order) && order[i] != expected {
			t.Errorf("expected order[%d] = %q, got %q", i, expected, order[i])
		}
	}
}

func TestGroupMultipleRoutes(t *testing.T) {
	g := NewRouter()
	getCalled := false
	postCalled := false

	subGroup := g.Group("/api")
	subGroup.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		getCalled = true
		w.WriteHeader(http.StatusOK)
	})
	subGroup.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		postCalled = true
		w.WriteHeader(http.StatusCreated)
	})

	// Test GET
	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)
	if !getCalled {
		t.Error("GET handler was not called")
	}

	// Test POST
	req = httptest.NewRequest("POST", "/api/users", nil)
	w = httptest.NewRecorder()
	g.ServeHTTP(w, req)
	if !postCalled {
		t.Error("POST handler was not called")
	}
}

func TestRouteParameters(t *testing.T) {
	tests := []struct {
		name          string
		pattern       string
		requestPath   string
		paramName     string
		expectedValue string
	}{
		{
			name:          "single parameter",
			pattern:       "/user/{id}",
			requestPath:   "/user/123",
			paramName:     "id",
			expectedValue: "123",
		},
		{
			name:          "parameter with prefix",
			pattern:       "/api/{name}",
			requestPath:   "/api/users",
			paramName:     "name",
			expectedValue: "users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewRouter()
			called := false
			var capturedValue string

			g.Get(tt.pattern, func(w http.ResponseWriter, r *http.Request) {
				called = true
				capturedValue = r.PathValue(tt.paramName)
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", tt.requestPath, nil)
			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)

			if !called {
				t.Error("handler was not called")
			}
			if capturedValue != tt.expectedValue {
				t.Errorf("expected parameter value %q, got %q", tt.expectedValue, capturedValue)
			}
		})
	}
}

func TestRouteWithMultipleParameters(t *testing.T) {
	g := NewRouter()
	called := false
	var capturedUserID, capturedPostID string

	g.Get("/user/{userId}/post/{postId}", func(w http.ResponseWriter, r *http.Request) {
		called = true
		capturedUserID = r.PathValue("userId")
		capturedPostID = r.PathValue("postId")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/user/123/post/456", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)

	if !called {
		t.Error("handler was not called")
	}
	if capturedUserID != "123" {
		t.Errorf("expected userId '123', got %q", capturedUserID)
	}
	if capturedPostID != "456" {
		t.Errorf("expected postId '456', got %q", capturedPostID)
	}
}

func TestRouteParameterInGroup(t *testing.T) {
	g := NewRouter()
	called := false
	var capturedID string

	apiGroup := g.Group("/api")
	apiGroup.Get("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		called = true
		capturedID = r.PathValue("id")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/user/123", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)

	if !called {
		t.Error("handler was not called")
	}
	if capturedID != "123" {
		t.Errorf("expected parameter value '123', got %q", capturedID)
	}
}

func TestRouteParameterWithDifferentValues(t *testing.T) {
	g := NewRouter()
	called := false
	var capturedValue string

	g.Get("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		called = true
		capturedValue = r.PathValue("id")
		w.WriteHeader(http.StatusOK)
	})

	// Test with different parameter values
	testCases := []string{"123", "456", "abc", "user-123"}
	for _, value := range testCases {
		called = false
		capturedValue = ""
		req := httptest.NewRequest("GET", "/user/"+value, nil)
		w := httptest.NewRecorder()
		g.ServeHTTP(w, req)

		if !called {
			t.Errorf("handler was not called for value %q", value)
		}
		if capturedValue != value {
			t.Errorf("expected parameter value %q, got %q", value, capturedValue)
		}
	}
}

func TestRoutePriorityExactOverParameter(t *testing.T) {
	tests := []struct {
		name          string
		registerOrder []string // "exact" or "param"
		exactPattern  string
		paramPattern  string
		requestPath   string
		expectExact   bool
		expectParam   bool
	}{
		{
			name:          "parameter first, exact second",
			registerOrder: []string{"param", "exact"},
			exactPattern:  "/user/123",
			paramPattern:  "/user/{id}",
			requestPath:   "/user/123",
			expectExact:   true,
			expectParam:   false,
		},
		{
			name:          "exact first, parameter second",
			registerOrder: []string{"exact", "param"},
			exactPattern:  "/user/123",
			paramPattern:  "/user/{id}",
			requestPath:   "/user/123",
			expectExact:   true,
			expectParam:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewRouter()
			exactCalled := false
			paramCalled := false

			for _, order := range tt.registerOrder {
				if order == "exact" {
					g.Get(tt.exactPattern, func(w http.ResponseWriter, r *http.Request) {
						exactCalled = true
						w.WriteHeader(http.StatusOK)
					})
				} else {
					g.Get(tt.paramPattern, func(w http.ResponseWriter, r *http.Request) {
						paramCalled = true
						w.WriteHeader(http.StatusOK)
					})
				}
			}

			req := httptest.NewRequest("GET", tt.requestPath, nil)
			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)

			if tt.expectExact && !exactCalled {
				t.Error("exact route handler was not called")
			}
			if !tt.expectExact && exactCalled {
				t.Error("exact route handler should not be called")
			}
			if tt.expectParam && !paramCalled {
				t.Error("parameter route handler was not called")
			}
			if !tt.expectParam && paramCalled {
				t.Error("parameter route handler should not be called")
			}
		})
	}
}

func TestRoutePriorityMoreSpecificOverLessSpecific(t *testing.T) {
	g := NewRouter()
	specificCalled := false
	generalCalled := false

	// Register general route first
	g.Get("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		generalCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Register more specific route second
	g.Get("/user/{id}/profile", func(w http.ResponseWriter, r *http.Request) {
		specificCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Request specific route
	req := httptest.NewRequest("GET", "/user/123/profile", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)

	// More specific route should be matched
	if !specificCalled {
		t.Error("specific route handler was not called")
	}
	if generalCalled {
		t.Error("general route handler should not be called for specific match")
	}
}

func TestRoutePriorityMultipleParameters(t *testing.T) {
	g := NewRouter()
	twoParamCalled := false
	oneParamCalled := false

	// Register single parameter route
	g.Get("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		oneParamCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Register two parameter route
	g.Get("/user/{userId}/post/{postId}", func(w http.ResponseWriter, r *http.Request) {
		twoParamCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Request two parameter route
	req := httptest.NewRequest("GET", "/user/123/post/456", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)

	// Two parameter route should be matched (more specific)
	if !twoParamCalled {
		t.Error("two parameter route handler was not called")
	}
	if oneParamCalled {
		t.Error("one parameter route handler should not be called")
	}
}

func TestRoutePriorityStaticOverParameter(t *testing.T) {
	g := NewRouter()
	staticCalled := false
	paramCalled := false

	// Register parameter route
	g.Get("/api/{name}", func(w http.ResponseWriter, r *http.Request) {
		paramCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Register static route
	g.Get("/api/users", func(w http.ResponseWriter, r *http.Request) {
		staticCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Request static route
	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)

	// Static route should be matched (higher priority)
	if !staticCalled {
		t.Error("static route handler was not called")
	}
	if paramCalled {
		t.Error("parameter route handler should not be called for static match")
	}
}

func TestRoutePriorityParameterFallback(t *testing.T) {
	g := NewRouter()
	paramCalled := false
	var capturedValue string

	// Register parameter route
	g.Get("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		paramCalled = true
		capturedValue = r.PathValue("id")
		w.WriteHeader(http.StatusOK)
	})

	// Request that doesn't match exact route should match parameter route
	req := httptest.NewRequest("GET", "/user/999", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)

	// Parameter route should be matched
	if !paramCalled {
		t.Error("parameter route handler was not called")
	}
	if capturedValue != "999" {
		t.Errorf("expected parameter value '999', got %q", capturedValue)
	}
}

func TestRouteWildcard(t *testing.T) {
	tests := []struct {
		name          string
		pattern       string
		requestPath   string
		expectedValue string
		setupGroup    func(*Router) *Router
	}{
		{
			name:          "root wildcard",
			pattern:       "/{pathname...}",
			requestPath:   "/api/users/123",
			expectedValue: "api/users/123",
			setupGroup:    func(g *Router) *Router { return g },
		},
		{
			name:          "single segment",
			pattern:       "/{pathname...}",
			requestPath:   "/test",
			expectedValue: "test",
			setupGroup:    func(g *Router) *Router { return g },
		},
		{
			name:          "with prefix",
			pattern:       "/api/{pathname...}",
			requestPath:   "/api/users/123/profile",
			expectedValue: "users/123/profile",
			setupGroup:    func(g *Router) *Router { return g },
		},
		{
			name:          "in group",
			pattern:       "/{pathname...}",
			requestPath:   "/api/users/123",
			expectedValue: "users/123",
			setupGroup:    func(g *Router) *Router { return g.Group("/api") },
		},
		{
			name:          "empty path",
			pattern:       "/{pathname...}",
			requestPath:   "/",
			expectedValue: "",
			setupGroup:    func(g *Router) *Router { return g },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewRouter()
			called := false
			var capturedPath string

			group := tt.setupGroup(g)
			group.Get(tt.pattern, func(w http.ResponseWriter, r *http.Request) {
				called = true
				capturedPath = r.PathValue("pathname")
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", tt.requestPath, nil)
			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)

			if !called {
				t.Error("handler was not called")
			}
			if capturedPath != tt.expectedValue {
				t.Errorf("expected path %q, got %q", tt.expectedValue, capturedPath)
			}
		})
	}
}

func TestRouteWildcardPriority(t *testing.T) {
	t.Run("parameter over wildcard", func(t *testing.T) {
		g := NewRouter()
		wildcardCalled := false
		paramCalled := false

		g.Get("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
			paramCalled = true
			w.WriteHeader(http.StatusOK)
		})
		g.Get("/{pathname...}", func(w http.ResponseWriter, r *http.Request) {
			wildcardCalled = true
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/user/123", nil)
		w := httptest.NewRecorder()
		g.ServeHTTP(w, req)

		if !paramCalled {
			t.Error("parameter route handler was not called")
		}
		if wildcardCalled {
			t.Error("wildcard route handler should not be called")
		}
	})

	t.Run("exact over wildcard", func(t *testing.T) {
		g := NewRouter()
		exactCalled := false
		wildcardCalled := false

		g.Get("/{pathname...}", func(w http.ResponseWriter, r *http.Request) {
			wildcardCalled = true
			w.WriteHeader(http.StatusOK)
		})
		g.Get("/api/users", func(w http.ResponseWriter, r *http.Request) {
			exactCalled = true
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/api/users", nil)
		w := httptest.NewRecorder()
		g.ServeHTTP(w, req)

		if !exactCalled {
			t.Error("exact route handler was not called")
		}
		if wildcardCalled {
			t.Error("wildcard route handler should not be called")
		}
	})

	t.Run("wildcard fallback", func(t *testing.T) {
		g := NewRouter()
		wildcardCalled := false
		var capturedPath string

		g.Get("/{pathname...}", func(w http.ResponseWriter, r *http.Request) {
			wildcardCalled = true
			capturedPath = r.PathValue("pathname")
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/some/deep/nested/path", nil)
		w := httptest.NewRecorder()
		g.ServeHTTP(w, req)

		if !wildcardCalled {
			t.Error("wildcard route handler was not called")
		}
		if capturedPath != "some/deep/nested/path" {
			t.Errorf("expected path 'some/deep/nested/path', got %q", capturedPath)
		}
	})
}

func TestRoutePriorityComplex(t *testing.T) {
	g := NewRouter()
	paramCalled := false
	exactUserIDCalled := false
	wildcardCalled := false
	var capturedID string
	var capturedPath string

	// Register routes in order:
	// 1. /user/{id} - parameter route
	g.Get("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		paramCalled = true
		capturedID = r.PathValue("id")
		w.WriteHeader(http.StatusOK)
	})

	// 2. /user/id/123 - exact route
	g.Get("/user/id/123", func(w http.ResponseWriter, r *http.Request) {
		exactUserIDCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// 3. /{pathname...} - wildcard route
	g.Get("/{pathname...}", func(w http.ResponseWriter, r *http.Request) {
		wildcardCalled = true
		capturedPath = r.PathValue("pathname")
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name              string
		requestPath       string
		expectParam       bool
		expectExactUserID bool
		expectWildcard    bool
		expectedID        string
		expectedPath      string
	}{
		{
			name:              "exact match /user/id/123 should match exact route, not parameter",
			requestPath:       "/user/id/123",
			expectExactUserID: true,
			expectParam:       false,
			expectWildcard:    false,
		},
		{
			name:           "parameter match /user/123 should match parameter route, not wildcard",
			requestPath:    "/user/123",
			expectParam:    true,
			expectWildcard: false,
			expectedID:     "123",
		},
		{
			name:           "wildcard fallback /some/other/path should match wildcard",
			requestPath:    "/some/other/path",
			expectWildcard: true,
			expectedPath:   "some/other/path",
		},
		{
			name:           "wildcard fallback /unknown should match wildcard",
			requestPath:    "/unknown",
			expectWildcard: true,
			expectedPath:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			paramCalled = false
			exactUserIDCalled = false
			wildcardCalled = false
			capturedID = ""
			capturedPath = ""

			req := httptest.NewRequest("GET", tt.requestPath, nil)
			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)

			if tt.expectParam && !paramCalled {
				t.Error("parameter route handler was not called")
			}
			if !tt.expectParam && paramCalled {
				t.Error("parameter route handler should not be called")
			}
			if tt.expectExactUserID && !exactUserIDCalled {
				t.Error("exact user ID route handler was not called")
			}
			if !tt.expectExactUserID && exactUserIDCalled {
				t.Error("exact user ID route handler should not be called")
			}
			if tt.expectWildcard && !wildcardCalled {
				t.Error("wildcard route handler was not called")
			}
			if !tt.expectWildcard && wildcardCalled {
				t.Error("wildcard route handler should not be called")
			}
			if tt.expectedID != "" && capturedID != tt.expectedID {
				t.Errorf("expected ID %q, got %q", tt.expectedID, capturedID)
			}
			if tt.expectedPath != "" && capturedPath != tt.expectedPath {
				t.Errorf("expected path %q, got %q", tt.expectedPath, capturedPath)
			}
		})
	}
}
