package middleware

import (
	"log"
	"net/http"
	"time"
)

func GetLoggingMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			log.Printf("Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
			next(w, r)
			duration := time.Since(start)
			log.Printf("Response time: %v", duration)
		}
	}
}
