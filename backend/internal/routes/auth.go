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

func AuthRoutes(router *gin.RouterGroup, log *zap.Logger, db *gorm.DB, cfg *config.Config) {
	svc := services.NewAuthService(log, db, cfg)
	controller := controllers.NewAuthController(log, svc, cfg.FrontendURL)

	auth := router.Group("/auth")
	auth.POST("/register", controller.CreateUser)
	auth.POST("/login", controller.Login)
	auth.POST("/refresh", controller.Refresh)
	auth.GET("/google/login", controller.HandleGoogleLogin)
	auth.GET("/google/callback", controller.HandleGoogleCallback)

	auth.POST("/verify-email", controller.VerifyEmail)
	auth.POST("/verify-email/resend", controller.ResendVerificationEmail)
	auth.POST("/password-reset/request", controller.RequestPasswordReset)
	auth.POST("/password-reset/confirm", controller.ConfirmPasswordReset)

	protected := auth.Group("")
	protected.Use(middlewares.AuthMiddleware(cfg))
	protected.GET("/me", controller.Me)
	protected.POST("/logout", controller.Logout)
}
