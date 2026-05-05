package services

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ayussh-2/config"
	"github.com/ayussh-2/internal/models"
	"github.com/ayussh-2/internal/utils"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	log *zap.Logger
	db  *gorm.DB
	cfg *config.Config
}
type UserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

type LoginResult struct {
	Response     LoginResponse
	RefreshToken string
}

type RefreshResult struct {
	Response     RefreshResponse
	RefreshToken string
}

type RegisterUserInput struct {
	Name     string
	Email    string
	Password string
}

func NewAuthService(log *zap.Logger, db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{log: log, db: db, cfg: cfg}
}

func (a *AuthService) RegisterUser(input RegisterUserInput) (*UserResponse, error) {
	user := models.User{
		Name:  input.Name,
		Email: input.Email,
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("failed to hash password", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot create user", err)
	}
	user.Password = string(hashedPassword)
	var existing models.User

	err = a.db.Where("email = ?", input.Email).First(&existing).Error
	if err == nil {
		return nil, utils.NewAppError(http.StatusConflict, "user already exists", nil)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		a.log.Error("failed to check existing email", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot create user", err)
	}

	if err := a.db.Create(&user).Error; err != nil {

		a.log.Error("failed to create user", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot create user", err)
	}

	resp := UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	return &resp, nil
}

func (a *AuthService) Login(input LoginInput) (*LoginResult, error) {
	var user models.User
	if err := a.db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAppError(http.StatusUnauthorized, "invalid credentials", nil)
		}
		a.log.Error("failed to fetch user for login", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot login", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, utils.NewAppError(http.StatusUnauthorized, "invalid credentials", nil)
	}

	accessToken, err := utils.GenerateAccessToken(a.cfg, user.ID, user.Email)
	if err != nil {
		a.log.Error("failed to generate access token", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot login", err)
	}
	refreshToken, refreshExpiry, err := utils.GenerateRefreshToken(a.cfg, user.ID)
	if err != nil {
		a.log.Error("failed to generate refresh token", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot login", err)
	}

	user.RefreshTokenHash = utils.HashToken(refreshToken)
	user.RefreshTokenExpiresAt = &refreshExpiry
	if err := a.db.Save(&user).Error; err != nil {
		a.log.Error("failed to save refresh token hash", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot login", err)
	}

	return &LoginResult{
		Response: LoginResponse{
			AccessToken: accessToken,
			User: UserResponse{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
			},
		},
		RefreshToken: refreshToken,
	}, nil
}

func (a *AuthService) Refresh(refreshToken string) (*RefreshResult, error) {
	claims, err := utils.ParseRefreshToken(a.cfg, refreshToken)
	if err != nil {
		return nil, utils.NewAppError(http.StatusUnauthorized, "invalid refresh token", nil)
	}

	var user models.User
	if err := a.db.First(&user, claims.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAppError(http.StatusUnauthorized, "invalid refresh token", nil)
		}
		a.log.Error("failed to fetch user on refresh", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot refresh token", err)
	}

	if user.RefreshTokenHash == "" || user.RefreshTokenExpiresAt == nil || user.RefreshTokenExpiresAt.Before(time.Now()) {
		return nil, utils.NewAppError(http.StatusUnauthorized, "refresh token expired", nil)
	}
	if user.RefreshTokenHash != utils.HashToken(refreshToken) {
		return nil, utils.NewAppError(http.StatusUnauthorized, "invalid refresh token", nil)
	}

	accessToken, err := utils.GenerateAccessToken(a.cfg, user.ID, user.Email)
	if err != nil {
		a.log.Error("failed to generate access token on refresh", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot refresh token", err)
	}
	newRefreshToken, refreshExpiry, err := utils.GenerateRefreshToken(a.cfg, user.ID)
	if err != nil {
		a.log.Error("failed to generate new refresh token", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot refresh token", err)
	}

	user.RefreshTokenHash = utils.HashToken(newRefreshToken)
	user.RefreshTokenExpiresAt = &refreshExpiry
	if err := a.db.Save(&user).Error; err != nil {
		a.log.Error("failed to rotate refresh token hash", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot refresh token", err)
	}

	return &RefreshResult{
		Response: RefreshResponse{
			AccessToken: accessToken,
		},
		RefreshToken: newRefreshToken,
	}, nil
}

func (a *AuthService) Logout(userID uint) error {
	err := a.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"refresh_token_hash":       "",
			"refresh_token_expires_at": nil,
		}).Error
	if err != nil {
		a.log.Error("failed to logout user", zap.Error(err))
		return utils.NewAppError(http.StatusInternalServerError, "cannot logout", err)
	}
	return nil
}

func (a *AuthService) RefreshCookieName() string {
	return a.cfg.RefreshCookie
}

func (a *AuthService) RefreshCookieMaxAgeSeconds() int {
	hours, err := strconv.Atoi(a.cfg.RefreshExpiryHrs)
	if err != nil {
		hours = 168
	}
	return hours * 3600
}
