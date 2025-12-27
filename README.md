# Mux
[![Go Reference](https://pkg.go.dev/badge/github.com/devilcove/mux?status.svg)](https://pkg.go.dev/github.com/devilcove/mux?tab=doc)  

a small idomatic http router and request multiplexer for go.
## Overview
provides a clean, expressive way to route HTTP requests in Go applications.  
* Route matching by path, host, query, headers, methods, and schemes
* Subrouters with shared conditions (prefix, middleware grouping)
* Middleware support (CORS, auth, logging, etc.)
* Static file serving and SPA-friendly routing
* Custom 404 and 405 handlers
* No dependencies
## Prerequisites
go version 1.24
## Running

Basic Example
```
package main

import (
	"fmt"
	"net/http"
	"io"

	"github.com/devilcove/mux"
)

func main() {
	r := mux.NewRouter(nil, middleware).NotFound(notFoundHandler).NotAllowed(notAllowedHandler)

	r.Get("/{$}", homeHandler)
	r.Get("/articles/{id}", articleHandler)

	// Serve static assets
	r.Static("/static", "static")

	r.Run(":8000")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to mux-powered app!")
}

func articleHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fmt.Fprintf(w, "Article ID: %s\n", id)
}

func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-tag", "true")
        next.ServeHTTP(w, r)
    })
}

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
    w.WriteHeader(http.StatusNotFound)
    io.WriteString(w, "custom not found")
}

func notAllowedHandler(w http.ResponseWriter, _ string, code int) {
    w.WriteHeader(code)
    io.WriteString(w, "Custom Method Not Allowed")
}
```

