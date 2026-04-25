package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimit defines a rate limiting rule.
type RateLimit struct {
	Max    int
	Window time.Duration
	Key    string // "email", "ip", or "userID"
}

// DefaultLimits defines per-endpoint rate limits.
var DefaultLimits = map[string]RateLimit{
	"/auth/otp/request":          {Max: 3, Window: 15 * time.Minute, Key: "email"},
	"/auth/otp/verify":           {Max: 5, Window: 5 * time.Minute, Key: "ip"},
	"/auth/token/refresh":        {Max: 30, Window: 1 * time.Minute, Key: "ip"},
	"/auth.v1.AuthService/GetMe": {Max: 60, Window: 1 * time.Minute, Key: "userID"},
	"default":                    {Max: 100, Window: 1 * time.Minute, Key: "ip"},
}

// RateLimiter provides Redis sliding window rate limiting.
type RateLimiter struct {
	rdb    *redis.Client
	limits map[string]RateLimit
}

// NewRateLimiter creates a new Redis-backed rate limiter.
func NewRateLimiter(rdb *redis.Client, limits map[string]RateLimit) *RateLimiter {
	if limits == nil {
		limits = DefaultLimits
	}
	return &RateLimiter{rdb: rdb, limits: limits}
}

// RateLimitMiddleware returns an HTTP middleware that enforces rate limits.
func (rl *RateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit, ok := rl.limits[r.URL.Path]
		if !ok {
			limit = rl.limits["default"]
		}

		identifier := extractIdentifier(r, limit.Key)
		key := fmt.Sprintf("rftw:rl:%s:%s", r.URL.Path, identifier)

		allowed, remaining, resetAt, err := rl.check(r.Context(), key, limit)
		if err != nil {
			// On Redis failure, allow the request (fail open for availability).
			next.ServeHTTP(w, r)
			return
		}

		// Set rate limit headers on all responses.
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit.Max))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		if !allowed {
			retryAfter := resetAt - time.Now().Unix()
			if retryAfter < 1 {
				retryAfter = 1
			}
			w.Header().Set("Retry-After", strconv.FormatInt(retryAfter, 10))
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// check uses a Redis sorted set sliding window to count requests.
func (rl *RateLimiter) check(ctx context.Context, key string, limit RateLimit) (allowed bool, remaining int, resetAt int64, err error) {
	now := time.Now()
	windowStart := now.Add(-limit.Window).UnixMicro()
	nowMicro := now.UnixMicro()
	resetAt = now.Add(limit.Window).Unix()

	pipe := rl.rdb.TxPipeline()
	// Remove entries outside the window.
	pipe.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(windowStart, 10))
	// Count current entries.
	countCmd := pipe.ZCard(ctx, key)
	// Add the current request.
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(nowMicro), Member: nowMicro})
	// Set key expiry to window duration.
	pipe.Expire(ctx, key, limit.Window)

	if _, err := pipe.Exec(ctx); err != nil {
		return true, limit.Max, resetAt, err
	}

	count := int(countCmd.Val())
	remaining = limit.Max - count - 1
	if remaining < 0 {
		remaining = 0
	}

	if count >= limit.Max {
		return false, 0, resetAt, nil
	}

	return true, remaining, resetAt, nil
}

func extractIdentifier(r *http.Request, keyType string) string {
	switch keyType {
	case "email":
		return r.FormValue("email")
	case "userID":
		if c := ClaimsFromContext(r.Context()); c != nil {
			return c.UserID
		}
		return extractIP(r)
	default: // "ip"
		return extractIP(r)
	}
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}
