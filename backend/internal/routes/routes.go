package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RegisterRoutes(server *gin.Engine, log *zap.Logger) {
	HealthRoutes(server, log)
}
