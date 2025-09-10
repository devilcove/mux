// Package mux is a small, idomatic http router
package mux

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// Middleware defines a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Router provides a chain of middlewares and routes.
type Router struct {
	*http.ServeMux
	*slog.Logger

	chain http.Handler
}

// DefaultRouter creates a new Router using the default ServeMux.
func DefaultRouter() *Router {
	logger = slog.New(slog.DiscardHandler)
	mux := http.NewServeMux()
	return &Router{
		ServeMux: mux,
		Logger:   logger,
		chain:    mux,
	}
}

// NewRouter creates a new Router with the given middleware applied.
func NewRouter(l *slog.Logger, middleware ...Middleware) *Router {
	r := DefaultRouter()
	if l != nil {
		r.Logger = l
		logger = l
	}
	r.Use(middleware...)
	return r
}

// Group creates a sub-router for the given prefix and applies middleware to it.
func (router *Router) Group(prefix string, middlewares ...Middleware) *Router {
	for _, m := range middlewares {
		if m == nil {
			panic("Router.Group: middleware cannot be nil")
		}
	}

	subRouter := DefaultRouter()
	subRouter.Use(middlewares...)
	router.Handle(prefix+"/", http.StripPrefix(prefix, subRouter))
	return subRouter
}

// Use adds a chain of middlewares to the router.
func (router *Router) Use(middlewares ...Middleware) {
	for _, m := range middlewares {
		router.chain = m(router.chain)
	}
}

// All registers the handler for all methods on given pattern.
func (router *Router) All(pattern string, handler http.HandlerFunc) {
	router.HandleFunc(pattern, handler)
}

// Post registers the handler for post requests on given pattern.
func (router *Router) Post(pattern string, handler http.HandlerFunc) {
	router.HandleFunc("POST\t"+pattern, handler)
}

// Get registers the handler for get requests on given pattern.
func (router *Router) Get(pattern string, handler http.HandlerFunc) {
	router.HandleFunc("GET\t"+pattern, handler)
}

// Delete registers the handler for delete requests on given pattern.
func (router *Router) Delete(pattern string, handler http.HandlerFunc) {
	router.HandleFunc("DELETE\t"+pattern, handler)
}

// Put registers the handler for Put requests on given pattern.
func (router *Router) Put(pattern string, handler http.HandlerFunc) {
	router.HandleFunc("PUT\t"+pattern, handler)
}

// Delete registers the handler for patch requests on given pattern.
func (router *Router) Patch(pattern string, handler http.HandlerFunc) {
	router.HandleFunc("PATCH\t"+pattern, handler)
}

// ServeHTTP implements the http.Handler interface.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.chain.ServeHTTP(w, r)
}

// Static registers the handle to serve static files.
func (router *Router) Static(pattern, dir string) {
	if !strings.HasSuffix(pattern, "/") {
		pattern += "/"
	}
	router.Handle(pattern, http.StripPrefix(pattern, http.FileServer(http.Dir(dir))))
}

// ServeFile registers a ServeFile handler.
func (router *Router) ServeFile(pattern, file string) {
	router.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, file)
	})
}

// Run starts the HTTP server and logs any error that occurs.
func (router *Router) Run(addr string) {
	server := http.Server{
		Addr:              addr,
		ReadHeaderTimeout: time.Second,
		Handler:           router,
	}
	router.Info("Starting server:", "Address", addr)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		router.Error("Router.Run: failed to start server: ", "error", err)
	}
}
