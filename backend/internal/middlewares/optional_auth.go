package middlewares

import (
	"github.com/ayussh-2/config"
	"github.com/ayussh-2/internal/utils"
	"github.com/gin-gonic/gin"
)

func OptionalAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := readAccessToken(c, cfg.AccessCookie)
		if token == "" {
			c.Next()
			return
		}

		claims, err := utils.ParseAccessToken(cfg, token)
		if err != nil {
			c.Next()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}
