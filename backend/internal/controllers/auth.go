package controllers

import (
	"errors"
	"net/http"

	"github.com/ayussh-2/internal/services"
	"github.com/ayussh-2/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthController struct {
	log *zap.Logger
	svc *services.AuthService
}

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=6,max=255"`
}

func NewAuthController(log *zap.Logger, svc *services.AuthService) *AuthController {
	return &AuthController{
		log: log,
		svc: svc,
	}
}

func (ac *AuthController) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.log.Error("failed to parse request body", zap.Error(err))
		utils.ValidationFail(c, "Validation Failed", utils.ValidationErrors(err))
		return
	}

	resp, err := ac.svc.RegisterUser(services.RegisterUserInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			utils.Fail(c, appErr.Status, appErr.Message)
			return
		}
		utils.Fail(c, http.StatusInternalServerError, "cannot create user")
		return
	}

	utils.Success(c, http.StatusCreated, "user created", resp)
}
