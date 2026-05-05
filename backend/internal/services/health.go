package services

import (
	"net/http"

	"github.com/ayussh-2/internal/utils"
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
	utils.Success(c, http.StatusOK, "service healthy", gin.H{"status": "ok"})
}

func (h *HealthService) Home(c *gin.Context) {
	utils.Success(c, http.StatusOK, "welcome", gin.H{"msg": "Welcome to something API"})
}
