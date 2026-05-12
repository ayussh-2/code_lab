package services

import (
	"context"
	"net/http"
	"time"

	"github.com/ayussh-2/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HealthService struct {
	log    *zap.Logger
	checks []DependencyCheck
}

type DependencyCheck struct {
	Name  string
	Check func(context.Context) error
}

func NewHealthService(log *zap.Logger, checks []DependencyCheck) *HealthService {
	return &HealthService{log: log, checks: checks}
}

func (h *HealthService) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	status := "ok"
	httpStatus := http.StatusOK
	dependencies := gin.H{}

	for _, dep := range h.checks {
		if err := dep.Check(ctx); err != nil {
			status = "degraded"
			httpStatus = http.StatusServiceUnavailable
			dependencies[dep.Name] = gin.H{
				"status": "down",
				"error":  err.Error(),
			}
			continue
		}
		dependencies[dep.Name] = gin.H{"status": "ok"}
	}

	h.log.Info("health check hit", zap.String("status", status))
	utils.Success(c, httpStatus, "service health", gin.H{
		"status":       status,
		"dependencies": dependencies,
	})
}
