package middlewares

import (
	"net/http"
	"strings"

	"github.com/ayussh-2/config"
	"github.com/ayussh-2/internal/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := readAccessToken(c, cfg.AccessCookie)
		if token == "" {
			utils.Fail(c, http.StatusUnauthorized, "missing credentials")
			c.Abort()
			return
		}

		claims, err := utils.ParseAccessToken(cfg, token)
		if err != nil {
			utils.Fail(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func readAccessToken(c *gin.Context, cookieName string) string {
	if cookie, err := c.Cookie(cookieName); err == nil && cookie != "" {
		return cookie
	}

	header := c.GetHeader("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		roleVal, ok := c.Get("role")
		if !ok {
			utils.Fail(c, http.StatusForbidden, "missing role in token")
			c.Abort()
			return
		}

		role, ok := roleVal.(string)
		if !ok {
			utils.Fail(c, http.StatusForbidden, "invalid role in token")
			c.Abort()
			return
		}

		if _, exists := allowed[role]; !exists {
			utils.Fail(c, http.StatusForbidden, "insufficient role permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}
