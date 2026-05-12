package routes

import (
	"github.com/ayussh-2/config"
	"github.com/ayussh-2/internal/controllers"
	"github.com/ayussh-2/internal/middlewares"
	"github.com/ayussh-2/internal/sandbox"
	"github.com/ayussh-2/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func SubmissionRoutes(router *gin.RouterGroup, log *zap.Logger, db *gorm.DB, cfg *config.Config, runner sandbox.Runner, publisher services.JudgePublisher) {
	svc := services.NewSubmissionService(log, db, cfg, runner, publisher)
	ctrl := controllers.NewSubmissionController(log, svc)

	g := router.Group("/submissions")
	g.Use(middlewares.AuthMiddleware(cfg))
	g.POST("", middlewares.UserRateLimit(cfg.SubmissionRatePerSec, cfg.SubmissionMaxPending), ctrl.CreateSubmission)
	g.GET("/:id", ctrl.GetSubmission)
	g.GET("", ctrl.ListSubmissions)
}
