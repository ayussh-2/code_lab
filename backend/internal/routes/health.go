package routes

import (
	"github.com/ayussh-2/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func HealthRoutes(router *gin.Engine, log *zap.Logger) {
	svc := services.NewHealthService(log)
	router.GET("/health", svc.HealthCheck)
	router.GET("/", svc.Home)
}
