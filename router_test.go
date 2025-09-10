package mux

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
	expectedLog := "GET example.com /ping 192.0.2.1"
	expectedForwardedLog := "GET example.com /ping 192.168.0.1"

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			next.ServeHTTP(w, r)
		})
	}

	buf := new(bytes.Buffer)
	logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{}))
	// NewLogger(logger)
	router := NewRouter(logger, middleware, Logger)
	router.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "pong")
	})

	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("User-Agent", "Go-Test")
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
	if !strings.Contains(buf.String(), expectedLog) {
		t.Errorf("Expected '%s', got '%s'", expectedLog, buf.String())
	}

	req.Header.Set("X-Forwarded-For", "192.168.0.1")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	body, _ = io.ReadAll(w.Result().Body)
	if string(body) != "pong" {
		t.Errorf("Expected 'pong', got '%s'", string(body))
	}
	if !strings.Contains(buf.String(), expectedForwardedLog) {
		t.Errorf("Expected '%s', got '%s'", expectedLog, buf.String())
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

func TestRouterRun(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{}))
	router := NewRouter(logger)
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "test response")
	})

	// start server
	go router.Run(":8000")
	time.Sleep(time.Millisecond * 10)
	line, err := buf.ReadString(0x0a)
	if err != nil {
		t.Error("read log buffer", err)
	}
	if !strings.Contains(line, "Starting server:") {
		t.Errorf("Expected 'Starting Server', got '%s'", buf)
	}
	line, err = buf.ReadString(0x0a)
	if err == nil || line != "" {
		t.Error("unexpected log message", line)
	}

	// start server again, should err with address already in use
	router.Run(":8000")
	line, err = buf.ReadString(0x0a)
	if err != nil {
		t.Error("read log buffer", err)
	}
	line, err = buf.ReadString(0x0a)
	if err != nil {
		t.Error("read log buffer", err)
	}
	if !strings.Contains(line, "failed to start server") {
		t.Error("server started but should not have")
	}
}

func TestStaticFiles(t *testing.T) {
	router := DefaultRouter()
	router.Static("/files", "example/static")
	router.ServeFile("/hello", "example/static/hello.txt")

	// get dir
	req := httptest.NewRequest("GET", "/files/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	body, err := io.ReadAll(w.Result().Body)
	if err != nil {
		t.Error("error reading body", err)
	}
	if !strings.Contains(string(body), "hello.txt") {
		t.Error("wrong response", string(body))
	}

	// get file from dir
	req = httptest.NewRequest("GET", "/files/hello.txt", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	body, err = io.ReadAll(w.Result().Body)
	if err != nil {
		t.Error("error reading body", err)
	}
	if !strings.Contains(string(body), "hello world") {
		t.Error("wrong response", string(body))
	}

	// get file directly
	req = httptest.NewRequest("GET", "/hello", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	body, err = io.ReadAll(w.Result().Body)
	if err != nil {
		t.Error("error reading body", err)
	}
	if !strings.Contains(string(body), "hello world") {
		t.Error("wrong response", string(body))
	}
}
