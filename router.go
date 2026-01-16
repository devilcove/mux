// Package mux is a small, idomatic http router
package mux

import (
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"
)

// Middleware defines a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Router provides a chain of middlewares and routes.
type Router struct {
	*http.ServeMux

	chain      http.Handler
	methods    []string
	notFound   func(http.ResponseWriter, *http.Request)
	notAllowed func(http.ResponseWriter, string, int)
}

// defaultRouter creates a new Router using the default ServeMux.
func defaultRouter() *Router {
	mux := http.NewServeMux()
	router := &Router{
		ServeMux:   mux,
		chain:      mux,
		methods:    []string{},
		notFound:   http.NotFound,
		notAllowed: http.Error,
	}
	// set up
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		_, current := mux.Handler(r)
		var allowed []string
		for _, method := range router.methods {
			// If we find a pattern that's different from the pattern for the
			// current fallback handler then we know there are actually other handlers
			// that could match with a method change, so we should handle as
			// method not allowed
			r.Method = method
			if _, pattern := mux.Handler(r); pattern != current {
				allowed = append(allowed, method)
			}
		}
		r.Method = method
		if len(allowed) != 0 {
			w.Header().Set("Allow", strings.Join(allowed, ", "))
			router.notAllowed(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		// http.Error(w, "Custom Not Found", http.StatusNotFound)
		router.notFound(w, r)
	})
	return router
}

// NewRouter creates a new Router with the given middleware applied.
func NewRouter(middleware ...Middleware) *Router {
	router := defaultRouter()
	router.Use(middleware...)
	return router
}

// NotFound sets a custome not found handler.
func (router *Router) NotFound(h func(http.ResponseWriter, *http.Request)) *Router {
	router.notFound = http.HandlerFunc(h)
	return router
}

// NotAllowed sets a custom method not allowed error.
func (router *Router) NotAllowed(h func(http.ResponseWriter, string, int)) *Router {
	router.notAllowed = h
	return router
}

// Group creates a sub-router for the given prefix and applies middleware to it.
func (router *Router) Group(prefix string, middlewares ...Middleware) *Router {
	for _, m := range middlewares {
		if m == nil {
			panic("Router.Group: middleware cannot be nil")
		}
	}

	subRouter := defaultRouter()
	subRouter.notFound = router.notFound
	subRouter.notAllowed = router.notAllowed
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
	methods := allMethods()
	for _, method := range methods {
		router.addMethod(method)
	}
	router.HandleFunc(pattern, handler)
}

// Post registers the handler for post requests on given pattern.
func (router *Router) Post(pattern string, handler http.HandlerFunc) {
	router.addMethod(http.MethodPost)
	router.HandleFunc(http.MethodPost+"\t"+pattern, handler)
}

// Get registers the handler for get requests on given pattern.
func (router *Router) Get(pattern string, handler http.HandlerFunc) {
	router.addMethod(http.MethodGet)
	router.addMethod(http.MethodHead)
	router.HandleFunc(http.MethodGet+"\t"+pattern, handler)
}

// Delete registers the handler for delete requests on given pattern.
func (router *Router) Delete(pattern string, handler http.HandlerFunc) {
	router.addMethod(http.MethodDelete)
	router.HandleFunc(http.MethodDelete+"\t"+pattern, handler)
}

// Put registers the handler for Put requests on given pattern.
func (router *Router) Put(pattern string, handler http.HandlerFunc) {
	router.addMethod(http.MethodPut)
	router.HandleFunc(http.MethodPut+"\t"+pattern, handler)
}

// Patch registers the handler for patch requests on given pattern.
func (router *Router) Patch(pattern string, handler http.HandlerFunc) {
	router.addMethod(http.MethodPatch)
	router.HandleFunc(http.MethodPatch+"\t"+pattern, handler)
}

// CustomMethod registers a handler for a custom method request.
func (router *Router) CustomMethod(method, pattern string, handle http.HandlerFunc) {
	router.addMethod(method)
	router.HandleFunc(method+"\t"+pattern, handle)
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

// StaticFS registers the handle to serve static files from FS filesystem.
// ex.
// //go:embded images
// var content embed.FS
// router.StaticFS("/images/", content) .
func (router *Router) StaticFS(pattern string, fs fs.FS) {
	if !strings.HasSuffix(pattern, "/") {
		pattern += "/"
	}
	router.Handle(pattern, http.StripPrefix(pattern, http.FileServer(http.FS(fs))))
}

// ServeFile registers a ServeFile handler.
func (router *Router) ServeFile(pattern, file string) {
	router.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, file)
	})
}

// ServeFileFS registers a ServeFileFS handler.
func (router *Router) ServeFileFS(pattern, file string, fs fs.FS) {
	router.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, fs, file)
	})
}

// Run starts the HTTP server and logs any error that occurs.
func (router *Router) Run(addr string) {
	server := http.Server{
		Addr:              addr,
		ReadHeaderTimeout: time.Second,
		Handler:           router,
	}
	slog.Info("Starting server:", "Address", addr)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Router.Run: failed to start server: ", "error", err)
	}
}

func (router *Router) addMethod(method string) {
	if !slices.Contains(router.methods, method) {
		router.methods = append(router.methods, method)
	}
}

func allMethods() []string {
	return []string{
		http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace,
	}
}
