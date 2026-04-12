package utils

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateBucket struct {
	count       int
	windowStart time.Time
}

var (
	rateMu     sync.Mutex
	rateLimits = make(map[string]*rateBucket)
)

func RateLimitMiddleware() gin.HandlerFunc {
	max := 60
	if v := os.Getenv("RATE_LIMIT_MAX"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			max = n
		}
	}
	windowSec := 60
	if v := os.Getenv("RATE_LIMIT_WINDOW_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			windowSec = n
		}
	}
	window := time.Duration(windowSec) * time.Second

	return func(c *gin.Context) {
		key := "ip:" + c.ClientIP()
		if uid, ok := OptionalJWTUserID(c); ok {
			key = "user:" + uid
		}

		now := time.Now()
		rateMu.Lock()
		b, ok := rateLimits[key]
		if !ok || now.Sub(b.windowStart) >= window {
			rateLimits[key] = &rateBucket{count: 1, windowStart: now}
			rateMu.Unlock()
			c.Next()
			return
		}
		if b.count >= max {
			rateMu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		b.count++
		rateMu.Unlock()
		c.Next()
	}
}
