package ratelimiter

import (
	"net/http"
)

type RateLimiter interface {
	Allow(ip string) bool
}

type Middleware struct {
	limiter RateLimiter
}

func NewMiddleware(limiter RateLimiter) *Middleware {
	return &Middleware{limiter: limiter}
}

func (m *Middleware) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if !m.limiter.Allow(ip) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}