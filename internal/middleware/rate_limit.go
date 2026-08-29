package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/Ashoke15/AuthX/internal/ratelimit"
)

func RateLimitByIP(limiter *ratelimit.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow(clientIP(r)) {
				writeTooManyRequest(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

func writeTooManyRequest(w http.ResponseWriter) {
	w.Header().Set("content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(map[string]string{"error": "too many request please try again leter"})
}
