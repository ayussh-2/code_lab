package middlewares

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ayussh-2/internal/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type userLimiter struct {
	limiter      *rate.Limiter
	lastSeenUnix atomic.Int64
}

func UserRateLimit(rps float64, burst int) gin.HandlerFunc {
	if rps <= 0 {
		rps = 1
	}
	if burst <= 0 {
		burst = 1
	}

	var limiters sync.Map
	go cleanupLimiters(&limiters, 10*time.Minute)

	return func(c *gin.Context) {
		userID, ok := c.Get("userID")
		if !ok {
			utils.Fail(c, http.StatusUnauthorized, "missing user")
			c.Abort()
			return
		}

		now := time.Now()
		fresh := &userLimiter{limiter: rate.NewLimiter(rate.Limit(rps), burst)}
		fresh.lastSeenUnix.Store(now.UnixNano())
		value, _ := limiters.LoadOrStore(userID, fresh)
		entry := value.(*userLimiter)
		entry.lastSeenUnix.Store(now.UnixNano())

		if !entry.limiter.Allow() {
			utils.Fail(c, http.StatusTooManyRequests, "slow down")
			c.Abort()
			return
		}

		c.Next()
	}
}

func cleanupLimiters(limiters *sync.Map, ttl time.Duration) {
	ticker := time.NewTicker(ttl)
	defer ticker.Stop()

	for range ticker.C {
		cutoff := time.Now().Add(-ttl)
		limiters.Range(func(key, value any) bool {
			entry := value.(*userLimiter)
			if time.Unix(0, entry.lastSeenUnix.Load()).Before(cutoff) {
				limiters.Delete(key)
			}
			return true
		})
	}
}
