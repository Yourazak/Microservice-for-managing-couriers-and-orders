package middleware

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	limit    int
	enabled  bool
	mu       sync.Mutex
	window   time.Duration
	burst    int
	requests map[string][]time.Time
}

type RateLimiterConfig struct {
	RequestsPerSecond float64
	Burst             int
	Enabled           bool
}

func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	if !config.Enabled {
		return &RateLimiter{enabled: false}
	}

	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    int(config.RequestsPerSecond),
		window:   time.Second,
		burst:    config.Burst,
		enabled:  true,
	}
}
func (rl *RateLimiter) Allow(ip string) bool {
	if !rl.enabled {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	validRequests := []time.Time{}
	for _, t := range rl.requests[ip] {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}

	if len(validRequests) >= rl.limit {
		if rl.burst > 0 && len(validRequests) < rl.burst {

		} else {
			return false
		}
	}

	validRequests = append(validRequests, now)
	rl.requests[ip] = validRequests

	return true
}
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.enabled {
				next.ServeHTTP(w, r)
				return
			}

			ip := r.RemoteAddr
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				ip = xff
			}

			if !rl.Allow(ip) {
				http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
