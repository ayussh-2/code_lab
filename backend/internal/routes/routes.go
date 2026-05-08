package routes

import (
	"github.com/ayussh-2/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func RegisterRoutes(server *gin.Engine, log *zap.Logger, db *gorm.DB, cfg *config.Config) {

	HealthRoutes(server, log)
	api := server.Group("/api")
	AuthRoutes(api, log, db, cfg)
	ProblemRoutes(api, log, db, cfg)
}
