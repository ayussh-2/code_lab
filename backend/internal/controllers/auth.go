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

type LoginRequest struct {
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

func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.log.Error("failed to parse login request", zap.Error(err))
		utils.ValidationFail(c, "Validation Failed", utils.ValidationErrors(err))
		return
	}

	result, err := ac.svc.Login(services.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			utils.Fail(c, appErr.Status, appErr.Message)
			return
		}
		utils.Fail(c, http.StatusInternalServerError, "cannot login")
		return
	}

	c.SetCookie(
		ac.svc.RefreshCookieName(),
		result.RefreshToken,
		ac.svc.RefreshCookieMaxAgeSeconds(),
		"/",
		"",
		false,
		true,
	)

	utils.Success(c, http.StatusOK, "login successful", result.Response)
}

func (ac *AuthController) Me(c *gin.Context) {
	userID, _ := c.Get("userID")
	email, _ := c.Get("email")
	utils.Success(c, http.StatusOK, "authenticated user", gin.H{
		"user_id": userID,
		"email":   email,
	})
}

func (ac *AuthController) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(ac.svc.RefreshCookieName())
	if err != nil || refreshToken == "" {
		utils.Fail(c, http.StatusUnauthorized, "missing refresh token")
		return
	}

	result, err := ac.svc.Refresh(refreshToken)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			utils.Fail(c, appErr.Status, appErr.Message)
			return
		}
		utils.Fail(c, http.StatusInternalServerError, "cannot refresh token")
		return
	}

	c.SetCookie(
		ac.svc.RefreshCookieName(),
		result.RefreshToken,
		ac.svc.RefreshCookieMaxAgeSeconds(),
		"/",
		"",
		false,
		true,
	)

	utils.Success(c, http.StatusOK, "token refreshed", result.Response)
}

func (ac *AuthController) Logout(c *gin.Context) {
	userIDVal, ok := c.Get("userID")
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := ac.svc.Logout(userID); err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			utils.Fail(c, appErr.Status, appErr.Message)
			return
		}
		utils.Fail(c, http.StatusInternalServerError, "cannot logout")
		return
	}

	c.SetCookie(ac.svc.RefreshCookieName(), "", -1, "/", "", false, true)
	utils.Success(c, http.StatusOK, "logout successful", nil)
}
