package routes

import (
	"github.com/ayussh-2/internal/controllers"
	"github.com/ayussh-2/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func AuthRoutes(router *gin.RouterGroup, log *zap.Logger, db *gorm.DB) {
	svc := services.NewAuthService(log, db)
	controller := controllers.NewAuthController(log, svc)
	router.POST("/users/register", controller.CreateUser)
}
