package middleware

import (
	"net/http"
	"shopify-lite/internal/utils"
	"sync"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	rate     rate.Limit
	burst    int
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if l, ok := rl.limiters[ip]; ok {
		return l
	}
	l := rate.NewLimiter(rl.rate, rl.burst)
	rl.limiters[ip] = l
	return l
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := utils.GetIP(r)
		if !rl.getLimiter(ip).Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		// NOTE: This map grows unboundedly. A production implementation
		// would use TTL-based eviction (e.g. with golang-lru or a cleanup goroutine).
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}
