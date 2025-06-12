# Mux
[![Go Reference](https://pkg.go.dev/badge/github.com/devilcove/mux?status.svg)](https://pkg.go.dev/github.com/devilcove/mux?tab=doc)  

a small idomatic http router and request multiplexer for go.
## Overview
provides a clean, expressive way to route HTTP requests in Go applications.  
* Route matching by path, host, query, headers, methods, and schemes
* Subrouters with shared conditions (prefix, middleware grouping)
* Middleware support (CORS, auth, logging, etc.)
* Static file serving and SPA-friendly routing
## Prerequisites
go version 1.23
## Running

Basic Example
```
package main

import (
	"fmt"
	"net/http"

	"github.com/devilcove/mux"
)

func main() {
	r := mux.NewRouter()

	r.Get("/{$}", homeHandler)
	r.Get("/articles/{id}", articleHandler)

	// Serve static assets
	r.Group("/static", static)

	r.Run(":8000")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to mux-powered app!")
}

func articleHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fmt.Fprintf(w, "Article ID: %s\n", id)
}

func static(next http.Handler) http.Handler {
	return http.FileServer(http.Dir("static"))
}
```
