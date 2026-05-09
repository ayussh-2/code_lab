package middlewares

import (
	"net/http"

	"github.com/ayussh-2/internal/utils"
	"github.com/gin-gonic/gin"
)

var unsafeMethods = map[string]struct{}{
	http.MethodPost:   {},
	http.MethodPut:    {},
	http.MethodPatch:  {},
	http.MethodDelete: {},
}

func CSRFOriginCheck(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, unsafe := unsafeMethods[c.Request.Method]; !unsafe {
			c.Next()
			return
		}

		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = c.GetHeader("Referer")
		}

		if origin == "" || !startsWith(origin, allowedOrigin) {
			utils.Fail(c, http.StatusForbidden, "invalid request origin")
			c.Abort()
			return
		}

		c.Next()
	}
}

func startsWith(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return s[:len(prefix)] == prefix
}
