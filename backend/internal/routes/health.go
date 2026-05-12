package routes

import (
	"context"
	"errors"

	"github.com/ayussh-2/internal/queue"
	"github.com/ayussh-2/internal/services"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func HealthRoutes(router *gin.Engine, log *zap.Logger, db *gorm.DB, queueClient *queue.Client, dockerClient *client.Client) {
	svc := services.NewHealthService(log, []services.DependencyCheck{
		{
			Name: "database",
			Check: func(ctx context.Context) error {
				sqlDB, err := db.DB()
				if err != nil {
					return err
				}
				return sqlDB.PingContext(ctx)
			},
		},
		{
			Name: "nats",
			Check: func(ctx context.Context) error {
				if queueClient == nil || !queueClient.IsConnected() {
					return errors.New("nats disconnected")
				}
				return nil
			},
		},
		{
			Name: "docker",
			Check: func(ctx context.Context) error {
				if dockerClient == nil {
					return errors.New("docker client unavailable")
				}
				_, err := dockerClient.Ping(ctx)
				return err
			},
		},
	})
	router.GET("/health", svc.HealthCheck)
}
