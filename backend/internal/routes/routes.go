package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func RegisterRoutes(server *gin.Engine, log *zap.Logger, db *gorm.DB) {
	
	HealthRoutes(server, log)
	api := server.Group("/api")
	AuthRoutes(api, log, db)
}
