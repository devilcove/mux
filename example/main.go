// Example programm
package main

import (
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/pprof"

	"github.com/devilcove/mux"
)

func main() {
	log.SetFlags(log.Lshortfile)

	router := mux.NewRouter(slog.Default(), mux.Logger)
	router.NotFound(notFound).NotAllowed(notAllowed)
	// router.Use(mux.Logger)
	router.Get("/{$}", page)
	router.Post("/hello", hello)
	router.HandleFunc("GET /junk", junk)
	router.Static("/pages/", "static")
	router.Static("/world", "static")
	router.ServeFile("/junk.txt", "static/hello.txt")
	router.Get("/debug/", pprof.Profile)
	// router.All("/", notFound)

	group1 := router.Group("/extra")
	group1.Use(extra)
	group1.All("/junk", junk)

	subGroup := group1.Group("/extra", empty)
	subGroup.Delete("/junk", junk)

	group2 := router.Group("/test", empty, extra, mux.Logger)
	// group2 := router.Group("/test", empty, extra)
	group2.Get("/{$}", page)
	group2.Get("/hello", hello)
	group3 := router.Group("/group3")
	group3.Get("/{$}", page)
	group3.Post("/hello", hello)
	group3.Delete("/hello", page)
	// router.All("/{path...}", router.NotFound)
	router.Run(":8080")
}

func page(w http.ResponseWriter, _ *http.Request) {
	log.Println("main page")
	io.WriteString(w, "main page")
}

func junk(w http.ResponseWriter, _ *http.Request) {
	log.Println("junk page")
	w.Write([]byte("junk page"))
}

func hello(w http.ResponseWriter, _ *http.Request) {
	log.Println("hello page")
	w.Write([]byte("hello world"))
}

func extra(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("extra middleware")
		next.ServeHTTP(w, r)
	})
}

func empty(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("empty middleware")
		next.ServeHTTP(w, r)
	})
}

func notFound(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	io.WriteString(w, "<!DOCTYPE html><div style=\"font-family: 'Bush Script MT', cursive;"+
		"font-size:xxx-large; text-align:center; margin:auto;\">"+
		"<p>This is not the page you are looking for ... </p>"+
		"<p>Go about your business</p>"+
		"<p>Move along</p></div>")
}

func notAllowed(w http.ResponseWriter, _ string, code int) {
	w.WriteHeader(code)
	io.WriteString(w, "Custom Method Not Allowed")
}
