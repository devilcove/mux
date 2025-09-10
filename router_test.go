package mux

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultRouter(t *testing.T) {
	router := DefaultRouter()
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "default router")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if string(body) != "default router" {
		t.Errorf("Expected 'default router', got '%s'", string(body))
	}
}

func TestMiddlewareExecution(t *testing.T) {
	called := false

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			next.ServeHTTP(w, r)
		})
	}

	router := NewRouter(slog.Default(), middleware, Logger)
	router.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "pong")
	})

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if !called {
		t.Errorf("Middleware was not called")
	}

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "pong" {
		t.Errorf("Expected 'pong', got '%s'", string(body))
	}
}

func TestGroupRouting(t *testing.T) {
	router := DefaultRouter()

	subRouter := router.Group("/api", func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Group", "called")
			next.ServeHTTP(w, r)
		})
	})

	subRouter.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, r.Header.Get("X-Group"))
	})

	req := httptest.NewRequest("GET", "/api/hello", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if string(body) != "called" {
		t.Errorf("Expected 'called' from middleware header, got '%s'", string(body))
	}
}

func TestNilMiddlewarePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when passing nil middleware to Group")
		}
	}()

	router := DefaultRouter()
	router.Group("/should-panic", nil)
}

func TestChainedMiddlewareOrder(t *testing.T) {
	var trace []string

	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			trace = append(trace, "m1")
			t.Log("m1")
			next.ServeHTTP(w, r)
		})
	}
	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			trace = append(trace, "m2")
			t.Log("m2")
			next.ServeHTTP(w, r)
		})
	}

	router := DefaultRouter()
	router.Use(m1, m2)

	router.HandleFunc("/chain", func(w http.ResponseWriter, r *http.Request) {
		trace = append(trace, "handler")
		t.Log("handler")
		io.WriteString(w, "done")
	})

	req := httptest.NewRequest("GET", "/chain", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	expected := []string{"m2", "m1", "handler"}
	t.Log(trace)

	for i, v := range expected {
		if trace[i] != v {
			t.Errorf("Expected '%s' at position %d, got '%s'", v, i, trace[i])
		}
	}
}

func TestRouter_All(t *testing.T) {
	router := DefaultRouter()
	router.All("/test", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "default router")
	})
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "default router" {
		t.Errorf("Expected 'default router', got '%s'", string(body))
	}
	req = httptest.NewRequest("POST", "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)
	if string(body) != "default router" {
		t.Errorf("Expected 'default router', got '%s'", string(body))
	}
}

func TestRouter_Get(t *testing.T) {
	router := DefaultRouter()
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "default router")
	})
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "default router" {
		t.Errorf("Expected 'default router', got '%s'", string(body))
	}
}

func TestRouter_Post(t *testing.T) {
	router := DefaultRouter()
	router.Post("/test", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "default router")
	})
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "default router" {
		t.Errorf("Expected 'default router', got '%s'", string(body))
	}
}

func TestRouter_Delete(t *testing.T) {
	router := DefaultRouter()
	router.Delete("/test", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "default router")
	})
	req := httptest.NewRequest(http.MethodDelete, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "default router" {
		t.Errorf("Expected 'default router', got '%s'", string(body))
	}
}

func TestRouter_Put(t *testing.T) {
	router := DefaultRouter()
	router.Put("/test", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "default router")
	})
	req := httptest.NewRequest(http.MethodPut, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "default router" {
		t.Errorf("Expected 'default router', got '%s'", string(body))
	}
}

func TestRouter_Patch(t *testing.T) {
	router := DefaultRouter()
	router.Patch("/test", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "default router")
	})
	req := httptest.NewRequest(http.MethodPatch, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "default router" {
		t.Errorf("Expected 'default router', got '%s'", string(body))
	}
}
