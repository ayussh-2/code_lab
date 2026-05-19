package routes

import (
	"github.com/ayussh-2/config"
	"github.com/ayussh-2/internal/controllers"
	"github.com/ayussh-2/internal/middlewares"
	"github.com/ayussh-2/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func ProblemRoutes(router *gin.RouterGroup, log *zap.Logger, db *gorm.DB, cfg *config.Config) {
	svc := services.NewProblemService(log, db, cfg)
	controller := controllers.NewProblemController(log, svc)

	problem := router.Group("/problems")
	problem.GET("/topics", controller.ListTopics)
	problem.GET("", middlewares.OptionalAuthMiddleware(cfg), controller.ListProblems)
	problem.GET("/:slug", middlewares.OptionalAuthMiddleware(cfg), controller.GetProblemBySlug)

	protected := problem.Group("")
	protected.Use(middlewares.AuthMiddleware(cfg), middlewares.RequireRoles("admin", "problem_setter"))
	protected.POST("/topics", controller.CreateTopic)
	protected.POST("/topics/bulk", controller.BulkCreateTopics)
	protected.POST("", controller.CreateProblem)
	protected.POST("/bulk", controller.BulkCreateProblems)
	protected.PATCH("/:slug", controller.UpdateProblem)
	protected.DELETE("/:slug", controller.DeleteProblem)
	protected.GET("/:slug/hidden-test-cases", controller.ListHiddenTestCases)
	protected.PUT("/:slug/hidden-test-cases", controller.ReplaceHiddenTestCases)
}
