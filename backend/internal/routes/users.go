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

func UserRoutes(router *gin.RouterGroup, log *zap.Logger, db *gorm.DB, cfg *config.Config) {
	svc := services.NewUserService(log, db, cfg)
	controller := controllers.NewUserController(log, svc)

	users := router.Group("/users")
	users.GET("/:username", controller.GetPublicProfile)
	users.GET("/:username/stats", controller.GetStats)
	users.GET("/:username/rating-history", controller.GetRatingHistory)
	users.GET("/:username/activity", controller.GetActivity)
	users.GET("/:username/submissions", controller.ListSubmissions)

	profile := router.Group("/profile")
	profile.Use(middlewares.AuthMiddleware(cfg))
	profile.GET("/me", controller.GetOwnProfile)
	profile.PATCH("/me", controller.UpdateOwnProfile)
}
