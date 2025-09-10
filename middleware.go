package mux

import (
	"log"
	"net/http"
	"strings"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter

	status int
}

// WriteHeader overrides std WriteHeader to save response code.
func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

// Logger is a logging middleware that logs useragent, RemoteAddr, Method, Host, Path and response.Status to stdlib log.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		rec := statusRecorder{w, http.StatusOK}
		next.ServeHTTP(&rec, r)
		remote := strings.Split(r.RemoteAddr, ":")[0]
		if r.Header.Get("X-Forwarded-For") != "" {
			remote = r.Header.Get("X-Forwarded-For")
		}
		log.Println(r.Method, r.Host, r.URL.Path, remote, rec.status, time.Since(now), r.UserAgent())
	})
}
