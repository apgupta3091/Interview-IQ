package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// limiterEntry holds a rate limiter and the last time it was accessed,
// used for periodic cleanup of idle entries.
type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// keyedRateLimiter manages per-key token-bucket limiters with background cleanup.
type keyedRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*limiterEntry
	r       rate.Limit
	b       int
}

func newKeyedRateLimiter(r rate.Limit, b int) *keyedRateLimiter {
	krl := &keyedRateLimiter{
		entries: make(map[string]*limiterEntry),
		r:       r,
		b:       b,
	}
	go krl.cleanup()
	return krl
}

// get returns (or creates) the rate limiter for the given key.
func (krl *keyedRateLimiter) get(key string) *rate.Limiter {
	krl.mu.Lock()
	defer krl.mu.Unlock()

	e, ok := krl.entries[key]
	if !ok {
		e = &limiterEntry{limiter: rate.NewLimiter(krl.r, krl.b)}
		krl.entries[key] = e
	}
	e.lastSeen = time.Now()
	return e.limiter
}

// cleanup removes entries that haven't been accessed in the last 5 minutes
// to prevent unbounded memory growth.
func (krl *keyedRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		krl.mu.Lock()
		for key, e := range krl.entries {
			if time.Since(e.lastSeen) > 5*time.Minute {
				delete(krl.entries, key)
			}
		}
		krl.mu.Unlock()
	}
}

// RateLimitByIP returns a middleware that enforces a per-IP token-bucket rate
// limit. r is the sustained rate (requests/second) and b is the burst size.
// Suitable for global application before any authentication.
//
// Example: RateLimitByIP(rate.Every(2*time.Second), 10) → 30 req/min, burst 10.
func RateLimitByIP(r rate.Limit, b int) func(http.Handler) http.Handler {
	krl := newKeyedRateLimiter(r, b)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ip := realIP(req)
			if !krl.get(ip).Allow() {
				http.Error(w, `{"error":"rate limit exceeded, please slow down"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}

// RateLimitByUser returns a middleware that enforces a per-user token-bucket
// rate limit. It must be placed after ClerkAuthenticate so that the userID is
// already present in the context. r is the sustained rate and b is the burst.
//
// Example: RateLimitByUser(rate.Every(time.Second), 20) → 60 req/min, burst 20.
func RateLimitByUser(r rate.Limit, b int) func(http.Handler) http.Handler {
	krl := newKeyedRateLimiter(r, b)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			userID := UserIDFromContext(req.Context())
			key := strconv.Itoa(userID)
			if !krl.get(key).Allow() {
				http.Error(w, `{"error":"rate limit exceeded, please slow down"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}

// realIP extracts the client IP from standard proxy headers, falling back to
// RemoteAddr. Chi's chimiddleware.RealIP does the same but we keep this local
// so the rate limiter has no extra dependency.
func realIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can be a comma-separated list; the first is the client.
		for i := 0; i < len(ip); i++ {
			if ip[i] == ',' {
				return ip[:i]
			}
		}
		return ip
	}
	// Strip the port from RemoteAddr (host:port).
	addr := r.RemoteAddr
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}
