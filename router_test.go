package mux

import (
	"bytes"
	"embed"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDefaultRouter(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("/test", func(w http.ResponseWriter, _ *http.Request) {
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

func TestLoggerMiddleware(t *testing.T) {
	expectedLog := "DELETE example.com/ 204 192.0.2.1"
	expectedForwardedLog := "DELETE example.com/ 204 192.168.0.1"
	buf := new(bytes.Buffer)
	logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{}))
	slog.SetDefault(logger)
	router := NewRouter(Logger)
	router.Delete("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	req.Header.Set("User-Agent", "Go-Test")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if !strings.Contains(buf.String(), expectedLog) {
		t.Errorf("Expected '%s', got '%s'", expectedLog, buf.String())
	}
	req.Header.Set("X-Forwarded-For", "192.168.0.1")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if !strings.Contains(buf.String(), expectedForwardedLog) {
		t.Errorf("Expected '%s', got '%s'", expectedLog, buf.String())
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
	router := NewRouter()
	router.Use(middleware)
	router.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, "pong")
	})
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
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
}

func TestGroupRouting(t *testing.T) {
	router := NewRouter()

	subRouter := router.Group("/api", func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Group", "called")
			next.ServeHTTP(w, r)
		})
	})

	subRouter.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, r.Header.Get("X-Group"))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
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

	router := NewRouter()
	router.Group("/should-panic", nil)
}

func TestChainedMiddlewareOrder(t *testing.T) {
	var trace []string

	mid1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			trace = append(trace, "m1")
			next.ServeHTTP(w, r)
		})
	}
	mid2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			trace = append(trace, "m2")
			next.ServeHTTP(w, r)
		})
	}

	router := NewRouter()
	router.Use(mid1, mid2)

	router.HandleFunc("/chain", func(w http.ResponseWriter, _ *http.Request) {
		trace = append(trace, "handler")
		io.WriteString(w, "done")
	})

	req := httptest.NewRequest(http.MethodGet, "/chain", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	expected := []string{"m2", "m1", "handler"}

	for i, v := range expected {
		if trace[i] != v {
			t.Errorf("Expected '%s' at position %d, got '%s'", v, i, trace[i])
		}
	}
}

func TestRouter_All(t *testing.T) {
	router := NewRouter()
	router.All("/test", func(w http.ResponseWriter, _ *http.Request) {
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
	req = httptest.NewRequest(http.MethodPost, "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)
	if string(body) != "default router" {
		t.Errorf("Expected 'default router', got '%s'", string(body))
	}
}

func TestRouter_Get(t *testing.T) {
	router := NewRouter()
	router.Get("/test", func(w http.ResponseWriter, _ *http.Request) {
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
	router := NewRouter()
	router.Post("/test", func(w http.ResponseWriter, _ *http.Request) {
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
	router := NewRouter()
	router.Delete("/test", func(w http.ResponseWriter, _ *http.Request) {
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
	router := NewRouter()
	router.Put("/test", func(w http.ResponseWriter, _ *http.Request) {
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
	router := NewRouter()
	router.Patch("/test", func(w http.ResponseWriter, _ *http.Request) {
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
	slog.SetDefault(logger)
	router := NewRouter()
	router.HandleFunc("/test", func(w http.ResponseWriter, _ *http.Request) {
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
	_, err = buf.ReadString(0x0a)
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
	router := NewRouter()
	router.Static("/files", "example/static")
	router.ServeFile("/hello", "example/static/hello.txt")

	// get dir
	req := httptest.NewRequest(http.MethodGet, "/files/", nil)
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
	req = httptest.NewRequest(http.MethodGet, "/files/hello.txt", nil)
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
	req = httptest.NewRequest(http.MethodGet, "/hello", nil)
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

//go:embed example
var content embed.FS

func TestStaticFilesFS(t *testing.T) {
	router := NewRouter()
	router.StaticFS("/files", content)
	router.ServeFileFS("/hello/", "example/static/hello.txt", content)

	// get dir
	t.Run("getDir", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		body, err := io.ReadAll(w.Result().Body)
		if err != nil {
			t.Error("error reading body", err)
		}
		if !strings.Contains(string(body), ">example/</a>") {
			t.Error("wrong response", string(body))
		}
	})

	// get file from dir
	t.Run("getFile", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/example/static/hello.txt", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		body, err := io.ReadAll(w.Result().Body)
		if err != nil {
			t.Error("error reading body", err)
		}
		if !strings.Contains(string(body), "hello world") {
			t.Error("wrong response", string(body))
		}
	})

	// get file directly
	t.Run("getFileDirect", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/hello/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		body, err := io.ReadAll(w.Result().Body)
		if err != nil {
			t.Error("error reading body", err)
		}
		if !strings.Contains(string(body), "hello world") {
			t.Error("wrong response", string(body))
			t.Log(w.Result().Header)
		}
	})
}

func TestErrorHandling(t *testing.T) {
	router := NewRouter().NotAllowed(
		func(w http.ResponseWriter, _ string, _ int) {
			http.Error(w, "Custom Method Not Allowed", http.StatusMethodNotAllowed)
		}).NotFound(
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			io.WriteString(w, "Custom Not Found")
		})
	router.Post("/{$}", func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, "hello")
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Error("expected", http.StatusMethodNotAllowed, "got", w.Result().StatusCode)
	}
	body, _ := io.ReadAll(w.Result().Body)
	if string(body) != "Custom Method Not Allowed\n" {
		t.Error("expected 'Custom Method Not Allowed' got", string(body))
	}
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Error("expected", http.StatusOK, "got", w.Result().StatusCode)
	}
	req = httptest.NewRequest(http.MethodGet, "/notfound", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusNotFound {
		t.Error("expected", http.StatusNotFound, "got", w.Result().StatusCode)
	}
	body, _ = io.ReadAll(w.Result().Body)
	if string(body) != "Custom Not Found" {
		t.Error("expected 'Custom Not Found' got", string(body))
	}
}

func TestCustomMethod(t *testing.T) {
	router := NewRouter()
	router.CustomMethod("UPDATE", "/{$}", func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, "custom method handler")
	})
	req := httptest.NewRequest("UPDATE", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("expected ", http.StatusOK, "got", resp.Status)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "custom method handler" {
		t.Errorf("Expected 'custom method handler', got '%s'", string(body))
	}
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Error("got", resp.Status, "expected", http.StatusMethodNotAllowed)
	}
	if resp.Header.Get("Allow") != "UPDATE" {
		t.Error("allowed header got", resp.Header.Get("Allowed"), "expected UPDATE")
	}
}
