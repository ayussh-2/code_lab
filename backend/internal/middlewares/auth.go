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
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Fail(c, http.StatusUnauthorized, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Fail(c, http.StatusUnauthorized, "invalid authorization header")
			c.Abort()
			return
		}

		claims, err := utils.ParseAccessToken(cfg, parts[1])
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
