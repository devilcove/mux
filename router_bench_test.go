package mux

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func dummyHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "ok")
}

func makeMiddleware(tag string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate middleware logic
			r.Header.Set("X-"+tag, "true")
			next.ServeHTTP(w, r)
		})
	}
}

func BenchmarkRouterBase(b *testing.B) {
	router := DefaultRouter()
	router.HandleFunc("/bench", dummyHandler)

	req := httptest.NewRequest("GET", "/bench", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRouterWithOneMiddleware(b *testing.B) {
	router := NewRouter(makeMiddleware("m1"))
	router.HandleFunc("/bench", dummyHandler)

	req := httptest.NewRequest("GET", "/bench", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRouterWithFiveMiddlewares(b *testing.B) {
	router := DefaultRouter()
	router.Use(
		makeMiddleware("m1"),
		makeMiddleware("m2"),
		makeMiddleware("m3"),
		makeMiddleware("m4"),
		makeMiddleware("m5"),
	)
	router.HandleFunc("/bench", dummyHandler)

	req := httptest.NewRequest("GET", "/bench", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRouterGroup(b *testing.B) {
	mainRouter := DefaultRouter()
	group := mainRouter.Group("/api", makeMiddleware("group"))

	group.HandleFunc("/bench", dummyHandler)

	req := httptest.NewRequest("GET", "/api/bench", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mainRouter.ServeHTTP(w, req)
	}
}
