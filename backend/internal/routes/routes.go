package routes

import (
	"github.com/ayussh-2/config"
	"github.com/ayussh-2/internal/queue"
	"github.com/ayussh-2/internal/sandbox"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func RegisterRoutes(server *gin.Engine, log *zap.Logger, db *gorm.DB, cfg *config.Config, runner sandbox.Runner, queueClient *queue.Client, dockerClient *client.Client) {
	HealthRoutes(server, log, db, queueClient, dockerClient)
	api := server.Group("/api")
	AuthRoutes(api, log, db, cfg)
	ProblemRoutes(api, log, db, cfg)
	SubmissionRoutes(api, log, db, cfg, runner, queueClient)
	UserRoutes(api, log, db, cfg)
}
