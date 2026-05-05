package services

import (
	"net/http"

	"github.com/ayussh-2/internal/models"
	"github.com/ayussh-2/internal/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthService struct {
	log *zap.Logger
	db  *gorm.DB
}
type UserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type RegisterUserInput struct {
	Name     string
	Email    string
	Password string
}

func NewAuthService(log *zap.Logger, db *gorm.DB) *AuthService {
	return &AuthService{log: log, db: db}
}

func (a *AuthService) RegisterUser(input RegisterUserInput) (*UserResponse, error) {
	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password,
	}
	var existing models.User

	err := a.db.Where("email = ?", input.Email).First(&existing).Error
	if err == nil {
		return nil, utils.NewAppError(http.StatusConflict, "user already exists", nil)
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
