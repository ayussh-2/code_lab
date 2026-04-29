package services

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HealthService struct {
	log *zap.Logger
}

func NewHealthService(log *zap.Logger) *HealthService {
	return &HealthService{log: log}
}

func (h *HealthService) HealthCheck(c *gin.Context) {
	h.log.Info("health check hit")
	c.JSON(200, gin.H{"status": "ok"})
}

func (h *HealthService) Home(c *gin.Context) {
	c.JSON(200, gin.H{"msg": "Welcome to something API"})
}
