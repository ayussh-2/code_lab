package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ayussh-2/internal/services"
	"github.com/ayussh-2/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserController struct {
	log *zap.Logger
	svc *services.UserService
}

type UpdateProfileRequest struct {
	Name      *string `json:"name" binding:"omitempty,min=2,max=100"`
	Username  *string `json:"username" binding:"omitempty,min=3,max=30"`
	AvatarURL *string `json:"avatar_url" binding:"omitempty,max=512"`
	Bio       *string `json:"bio" binding:"omitempty,max=500"`
}

func NewUserController(log *zap.Logger, svc *services.UserService) *UserController {
	return &UserController{log: log, svc: svc}
}

func (uc *UserController) GetPublicProfile(c *gin.Context) {
	username := c.Param("username")
	resp, err := uc.svc.GetPublicProfile(username)
	if err != nil {
		handleAppError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "profile fetched", resp)
}

func (uc *UserController) GetStats(c *gin.Context) {
	resp, err := uc.svc.GetStats(c.Param("username"))
	if err != nil {
		handleAppError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "stats fetched", resp)
}

func (uc *UserController) GetRatingHistory(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	resp, err := uc.svc.GetRatingHistory(c.Param("username"), limit)
	if err != nil {
		handleAppError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "rating history fetched", resp)
}

func (uc *UserController) GetActivity(c *gin.Context) {
	resp, err := uc.svc.GetActivityHeatmap(c.Param("username"))
	if err != nil {
		handleAppError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "activity fetched", resp)
}

func (uc *UserController) ListSubmissions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	resp, err := uc.svc.ListProfileSubmissions(c.Param("username"), limit)
	if err != nil {
		handleAppError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "submissions fetched", resp)
}

func (uc *UserController) GetOwnProfile(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	resp, err := uc.svc.GetOwnProfile(userID)
	if err != nil {
		handleAppError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "profile fetched", resp)
}

func (uc *UserController) UpdateOwnProfile(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationFail(c, "Validation Failed", utils.ValidationErrors(err))
		return
	}

	resp, err := uc.svc.UpdateProfile(userID, services.UpdateProfileInput{
		Name:      req.Name,
		Username:  req.Username,
		AvatarURL: req.AvatarURL,
		Bio:       req.Bio,
	})
	if err != nil {
		handleAppError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "profile updated", resp)
}

func currentUserID(c *gin.Context) (uint, bool) {
	userIDVal, ok := c.Get("userID")
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, "unauthorized")
		return 0, false
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		utils.Fail(c, http.StatusUnauthorized, "unauthorized")
		return 0, false
	}
	return userID, true
}

func handleAppError(c *gin.Context, err error) {
	var appErr *utils.AppError
	if errors.As(err, &appErr) {
		utils.Fail(c, appErr.Status, appErr.Message)
		return
	}
	utils.Fail(c, http.StatusInternalServerError, "request failed")
}
