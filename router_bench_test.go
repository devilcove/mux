package mux

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func dummyHandler(w http.ResponseWriter, _ *http.Request) {
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
	router := defaultRouter()
	router.HandleFunc("/bench", dummyHandler)

	req := httptest.NewRequest(http.MethodGet, "/bench", nil)
	w := httptest.NewRecorder()

	for b.Loop() {
		router.ServeHTTP(w, req)
	}
}

func Benchmark5Routes(b *testing.B) {
	router := defaultRouter()
	router.HandleFunc("/bench1", dummyHandler)
	router.HandleFunc("/bench2", dummyHandler)
	router.HandleFunc("/bench3", dummyHandler)
	router.HandleFunc("/bench4", dummyHandler)
	router.HandleFunc("/bench5", dummyHandler)

	req := httptest.NewRequest(http.MethodGet, "/bench2", nil)
	w := httptest.NewRecorder()

	for b.Loop() {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRouterWithOneMiddleware(b *testing.B) {
	router := NewRouter(makeMiddleware("m1"))
	router.HandleFunc("/bench", dummyHandler)

	req := httptest.NewRequest(http.MethodGet, "/bench", nil)
	w := httptest.NewRecorder()

	for b.Loop() {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRouterWithFiveMiddlewares(b *testing.B) {
	router := defaultRouter()
	router.Use(
		makeMiddleware("m1"),
		makeMiddleware("m2"),
		makeMiddleware("m3"),
		makeMiddleware("m4"),
		makeMiddleware("m5"),
	)
	router.HandleFunc("/bench", dummyHandler)

	req := httptest.NewRequest(http.MethodGet, "/bench", nil)
	w := httptest.NewRecorder()

	for b.Loop() {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRouterGroup(b *testing.B) {
	mainRouter := defaultRouter()
	group := mainRouter.Group("/api", makeMiddleware("group"))

	group.HandleFunc("/bench", dummyHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/bench", nil)
	w := httptest.NewRecorder()

	for b.Loop() {
		mainRouter.ServeHTTP(w, req)
	}
}

func BenchmarkLogginMiddleWare(b *testing.B) {
	router := NewRouter(Logger)
	router.HandleFunc("/bench", dummyHandler)
	req := httptest.NewRequest(http.MethodGet, "/bench", nil)
	w := httptest.NewRecorder()
	for b.Loop() {
		router.ServeHTTP(w, req)
	}
}

// func BenchmarkNotFoundMiddleware(b *testing.B) {
// 	router := NewRouter(slog.New(slog.DiscardHandler), NotFound)
// 	router.HandleFunc("/bench", dummyHandler)
// 	req := httptest.NewRequest(http.MethodGet, "/bench", nil)
// 	w := httptest.NewRecorder()
// 	b.ResetTimer()
// 	for range b.N {
// 		router.ServeHTTP(w, req)
// 	}
// }

// func BenchmarkLoggingAndNotFoundMiddleware(b *testing.B) {
// 	router := NewRouter(slog.New(slog.DiscardHandler), Logger, NotFound)
// 	router.HandleFunc("/bench", dummyHandler)
// 	req := httptest.NewRequest(http.MethodGet, "/bench", nil)
// 	w := httptest.NewRecorder()
// 	b.ResetTimer()
// 	for range b.N {
// 		router.ServeHTTP(w, req)
// 	}
// }
