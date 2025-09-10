// Example programm
package main

import (
	"io"
	"log"
	"log/slog"
	"net/http"

	"github.com/devilcove/mux"
)

func main() {
	log.SetFlags(log.Lshortfile)
	router := mux.NewRouter(slog.Default())
	router.Use(mux.Logger)
	router.Get("/{$}", page)
	router.Post("/hello", hello)
	router.HandleFunc("GET /junk", junk)
	router.Static("/pages/", "static")
	router.Static("/world", "static")
	router.ServeFile("/junk.txt", "static/hello.txt")
	router.All("/", notFound)
	group1 := router.Group("/extra")
	group1.Use(extra)
	group1.All("/junk", junk)
	subGroup := group1.Group("/extra", empty)
	subGroup.Delete("/junk", junk)
	group2 := router.Group("/test", empty, extra, mux.Logger)
	group2.Get("/{$}", page)
	group2.Get("/hello", hello)
	router.Run(":8080")
}

func notFound(w http.ResponseWriter, _ *http.Request) {
	log.Println("not found handlers")
	w.WriteHeader(http.StatusNotFound)
	io.WriteString(w, "This is not the page you are looking for ... \nGo about your business")
}

func page(w http.ResponseWriter, _ *http.Request) {
	log.Println("main page")
	w.Write([]byte("main page"))
}

func junk(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("junk page"))
}

func hello(w http.ResponseWriter, _ *http.Request) {
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
