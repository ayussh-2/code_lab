package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/ayussh-2/config"
	"github.com/ayussh-2/internal/models"
	"github.com/ayussh-2/internal/utils"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	otpCodeLength  = 6
	otpExpiry      = 10 * time.Minute
	otpMaxAttempts = 5
)

type AuthService struct {
	log *zap.Logger
	db  *gorm.DB
	cfg *config.Config
}
type UserResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio"`
	Rating    int    `json:"rating"`
	Role      string `json:"role"`
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginResponse struct {
	User UserResponse `json:"user"`
}

type RefreshResponse struct {
	User UserResponse `json:"user"`
}

type LoginResult struct {
	Response     LoginResponse
	AccessToken  string
	RefreshToken string
}

type RefreshResult struct {
	Response     RefreshResponse
	AccessToken  string
	RefreshToken string
}

type RegisterUserInput struct {
	Name     string
	Email    string
	Password string
}

type GoogleSignupInput struct {
	Name  string
	Email string
}



func NewAuthService(log *zap.Logger, db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{log: log, db: db, cfg: cfg}
}

func (a *AuthService) RegisterUser(input RegisterUserInput) (*UserResponse, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" {
		return nil, utils.NewAppError(http.StatusBadRequest, "email is required", nil)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("failed to hash password", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot create user", err)
	}

	var user models.User
	err = a.db.Where("email = ?", email).First(&user).Error
	switch {
	case err == nil:
		if user.EmailVerified {
			return nil, utils.NewAppError(http.StatusConflict, "user already exists", nil)
		}
		user.Name = input.Name
		user.Password = string(hashedPassword)
		if err := a.db.Save(&user).Error; err != nil {
			a.log.Error("failed to update unverified user", zap.Error(err))
			return nil, utils.NewAppError(http.StatusInternalServerError, "cannot create user", err)
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		user = models.User{
			Name:          input.Name,
			Email:         email,
			Password:      string(hashedPassword),
			Role:          "user",
			EmailVerified: false,
		}
		if err := a.db.Create(&user).Error; err != nil {
			a.log.Error("failed to create user", zap.Error(err))
			return nil, utils.NewAppError(http.StatusInternalServerError, "cannot create user", err)
		}
	default:
		a.log.Error("failed to check existing email", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot create user", err)
	}

	if err := a.issueAndSendOTP(&user.ID, user.Email, models.OTPPurposeEmailVerify); err != nil {
		return nil, err
	}

	a.ensureProfileDefaults(&user)

	resp := a.userResponse(&user)
	return &resp, nil
}

func (a *AuthService) RegisterGoogleUser(input GoogleSignupInput) (*UserResponse, error) {
	randomPassword := make([]byte, 24)
	if _, err := rand.Read(randomPassword); err != nil {
		a.log.Error("failed to generate random password for google user", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot create user", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(base64.RawURLEncoding.EncodeToString(randomPassword)), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("failed to hash password for google user", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot create user", err)
	}

	user := models.User{
		Name:          input.Name,
		Email:         strings.TrimSpace(strings.ToLower(input.Email)),
		Password:      string(hashedPassword),
		Role:          "user",
		EmailVerified: true,
	}

	if err := a.db.Create(&user).Error; err != nil {
		a.log.Error("failed to create google user", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot create user", err)
	}

	a.ensureProfileDefaults(&user)

	resp := a.userResponse(&user)
	return &resp, nil
}

func (a *AuthService) Login(input LoginInput, callbackEmail string, callbackName string) (*LoginResult, error) {
	var user models.User

	finalEmail := input.Email
	if callbackEmail != "" {
		finalEmail = callbackEmail
	}

	if err := a.db.Where("email = ?", finalEmail).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if callbackEmail != "" {
				if _, registerErr := a.RegisterGoogleUser(GoogleSignupInput{
					Name:  callbackName,
					Email: callbackEmail,
				}); registerErr != nil {
					return nil, registerErr
				}
				if loginAgainErr := a.db.Where("email = ?", finalEmail).First(&user).Error; loginAgainErr != nil {
					a.log.Error("failed to fetch google user after registration", zap.Error(loginAgainErr))
					return nil, utils.NewAppError(http.StatusInternalServerError, "cannot login", loginAgainErr)
				}
			} else {
				return nil, utils.NewAppError(http.StatusUnauthorized, "invalid credentials", nil)
			}
		} else {
			a.log.Error("failed to fetch user for login", zap.Error(err))
			return nil, utils.NewAppError(http.StatusInternalServerError, "cannot login", err)
		}
	}

	if callbackEmail == "" {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
			return nil, utils.NewAppError(http.StatusUnauthorized, "invalid credentials", nil)
		}
		if !user.EmailVerified {
			return nil, utils.NewAppError(http.StatusForbidden, "email not verified", nil)
		}
	}

	accessToken, err := utils.GenerateAccessToken(a.cfg, user.ID, user.Email, user.Role)
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

	a.ensureProfileDefaults(&user)

	return &LoginResult{
		Response: LoginResponse{
			User: a.userResponse(&user),
		},
		AccessToken:  accessToken,
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

	accessToken, err := utils.GenerateAccessToken(a.cfg, user.ID, user.Email, user.Role)
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

	a.ensureProfileDefaults(&user)

	return &RefreshResult{
		Response: RefreshResponse{
			User: a.userResponse(&user),
		},
		AccessToken:  accessToken,
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

func (a *AuthService) AccessCookieName() string {
	return a.cfg.AccessCookie
}

func (a *AuthService) AccessCookieMaxAgeSeconds() int {
	minutes, err := strconv.Atoi(a.cfg.JWTExpiryMinutes)
	if err != nil {
		minutes = 15
	}
	return minutes * 60
}

func (a *AuthService) IsSecureCookie() bool {
	return a.cfg.Env == "production"
}

func (a *AuthService) GetUserByID(id uint) (*UserResponse, error) {
	var user models.User
	if err := a.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAppError(http.StatusNotFound, "user not found", nil)
		}
		a.log.Error("failed to fetch user", zap.Error(err))
		return nil, utils.NewAppError(http.StatusInternalServerError, "cannot fetch user", err)
	}
	a.ensureProfileDefaults(&user)
	resp := a.userResponse(&user)
	return &resp, nil
}

func (a *AuthService) userResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Username:  user.Username,
		AvatarURL: user.AvatarURL,
		Bio:       user.Bio,
		Rating:    user.Rating,
		Role:      user.Role,
	}
}

func (a *AuthService) ensureProfileDefaults(user *models.User) {
	userSvc := NewUserService(a.log, a.db, a.cfg)
	if err := userSvc.EnsureUsername(user); err != nil {
		a.log.Warn("failed to ensure username", zap.Error(err), zap.Uint("user_id", user.ID))
		return
	}
	_ = a.db.First(user, user.ID).Error
	if user.Rating == 0 {
		user.Rating = 1500
		_ = a.db.Model(user).Update("rating", 1500).Error
	}
	_ = userSvc.SeedInitialRatingHistory(user.ID, user.Rating, user.CreatedAt)
}


func (a *AuthService) RequestPasswordResetOTP(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return utils.NewAppError(http.StatusBadRequest, "email is required", nil)
	}

	var user models.User
	if err := a.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			a.log.Info("password reset requested for unknown email", zap.String("email", email))
			return nil
		}
		a.log.Error("failed to lookup user for password reset", zap.Error(err))
		return utils.NewAppError(http.StatusInternalServerError, "cannot send reset code", err)
	}

	return a.issueAndSendOTP(&user.ID, email, models.OTPPurposePasswordReset)
}

func (a *AuthService) ConfirmPasswordReset(email, code, newPassword string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || code == "" || newPassword == "" {
		return utils.NewAppError(http.StatusBadRequest, "email, code and new password are required", nil)
	}
	if len(newPassword) < 6 {
		return utils.NewAppError(http.StatusBadRequest, "password must be at least 6 characters", nil)
	}

	if err := a.consumeOTP(email, code, models.OTPPurposePasswordReset); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("failed to hash new password", zap.Error(err))
		return utils.NewAppError(http.StatusInternalServerError, "cannot reset password", err)
	}

	res := a.db.Model(&models.User{}).
		Where("email = ?", email).
		Updates(map[string]any{
			"password":                 string(hashedPassword),
			"refresh_token_hash":       "",
			"refresh_token_expires_at": nil,
			"email_verified":           true,
		})
	if res.Error != nil {
		a.log.Error("failed to update password", zap.Error(res.Error))
		return utils.NewAppError(http.StatusInternalServerError, "cannot reset password", res.Error)
	}
	if res.RowsAffected == 0 {
		return utils.NewAppError(http.StatusNotFound, "user not found", nil)
	}

	return nil
}

func (a *AuthService) VerifyEmail(email, code string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || code == "" {
		return utils.NewAppError(http.StatusBadRequest, "email and code are required", nil)
	}

	if err := a.consumeOTP(email, code, models.OTPPurposeEmailVerify); err != nil {
		return err
	}

	res := a.db.Model(&models.User{}).
		Where("email = ?", email).
		Update("email_verified", true)
	if res.Error != nil {
		a.log.Error("failed to mark email verified", zap.Error(res.Error))
		return utils.NewAppError(http.StatusInternalServerError, "cannot verify email", res.Error)
	}
	if res.RowsAffected == 0 {
		return utils.NewAppError(http.StatusNotFound, "user not found", nil)
	}

	return nil
}

func (a *AuthService) ResendEmailVerificationOTP(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return utils.NewAppError(http.StatusBadRequest, "email is required", nil)
	}

	var user models.User
	if err := a.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			a.log.Info("verification resend for unknown email", zap.String("email", email))
			return nil
		}
		a.log.Error("failed to lookup user for verification resend", zap.Error(err))
		return utils.NewAppError(http.StatusInternalServerError, "cannot send verification code", err)
	}

	if user.EmailVerified {
		return nil
	}

	return a.issueAndSendOTP(&user.ID, email, models.OTPPurposeEmailVerify)
}

func (a *AuthService) issueAndSendOTP(userID *uint, email, purpose string) error {
	if err := a.db.Model(&models.OTP{}).
		Where("email = ? AND purpose = ? AND consumed_at IS NULL", email, purpose).
		Update("consumed_at", time.Now()).Error; err != nil {
		a.log.Error("failed to invalidate previous otps", zap.Error(err), zap.String("purpose", purpose))
		return utils.NewAppError(http.StatusInternalServerError, "cannot send code", err)
	}

	code, err := generateOTPCode(otpCodeLength)
	if err != nil {
		a.log.Error("failed to generate otp code", zap.Error(err))
		return utils.NewAppError(http.StatusInternalServerError, "cannot send code", err)
	}

	otp := models.OTP{
		UserID:    userID,
		Email:     email,
		CodeHash:  hashOTPCode(email, code),
		Purpose:   purpose,
		ExpiresAt: time.Now().Add(otpExpiry),
	}
	if err := a.db.Create(&otp).Error; err != nil {
		a.log.Error("failed to persist otp", zap.Error(err))
		return utils.NewAppError(http.StatusInternalServerError, "cannot send code", err)
	}

	if err := a.sendOTPEmail(email, code, purpose, int(otpExpiry/time.Minute)); err != nil {
		a.log.Error("failed to send otp email", zap.Error(err))
		return utils.NewAppError(http.StatusInternalServerError, "cannot send code", err)
	}

	return nil
}

func (a *AuthService) consumeOTP(email, code, purpose string) error {
	var otp models.OTP
	err := a.db.Where(
		"email = ? AND purpose = ? AND consumed_at IS NULL",
		email, purpose,
	).Order("created_at desc").First(&otp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.NewAppError(http.StatusUnauthorized, "invalid or expired code", nil)
		}
		a.log.Error("failed to lookup otp", zap.Error(err))
		return utils.NewAppError(http.StatusInternalServerError, "cannot verify code", err)
	}

	if time.Now().After(otp.ExpiresAt) {
		return utils.NewAppError(http.StatusUnauthorized, "invalid or expired code", nil)
	}
	if otp.Attempts >= otpMaxAttempts {
		now := time.Now()
		otp.ConsumedAt = &now
		_ = a.db.Save(&otp).Error
		return utils.NewAppError(http.StatusUnauthorized, "too many attempts, request a new code", nil)
	}

	if otp.CodeHash != hashOTPCode(email, strings.TrimSpace(code)) {
		otp.Attempts++
		if err := a.db.Save(&otp).Error; err != nil {
			a.log.Error("failed to increment otp attempts", zap.Error(err))
		}
		return utils.NewAppError(http.StatusUnauthorized, "invalid or expired code", nil)
	}

	now := time.Now()
	otp.ConsumedAt = &now
	if err := a.db.Save(&otp).Error; err != nil {
		a.log.Error("failed to mark otp consumed", zap.Error(err))
		return utils.NewAppError(http.StatusInternalServerError, "cannot verify code", err)
	}

	return nil
}

func (a *AuthService) sendOTPEmail(toEmail, code, purpose string, expiresInMinutes int) error {
	from := a.cfg.GmailFrom
	if from == "" {
		from = a.cfg.GmailUserName
	}

	auth := smtp.PlainAuth("", a.cfg.GmailUserName, a.cfg.GmailPassword, a.cfg.GmailHost)

	subject, body := otpEmailContent(purpose, code, expiresInMinutes)

	headers := []string{
		"From: " + from,
		"To: " + toEmail,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=\"utf-8\"",
		"",
		body,
	}
	msg := []byte(strings.Join(headers, "\r\n"))

	addr := a.cfg.GmailHost + ":" + a.cfg.GmailPort
	return smtp.SendMail(addr, auth, from, []string{toEmail}, msg)
}

func otpEmailContent(purpose, code string, expiresInMinutes int) (subject string, body string) {
	switch purpose {
	case models.OTPPurposeEmailVerify:
		subject = "Verify your CODE_LAB email"
		body = fmt.Sprintf(
			"Hi,\r\n\r\n"+
				"Welcome to CODE_LAB! Use the code below to verify your email address.\r\n\r\n"+
				"    %s\r\n\r\n"+
				"This code expires in %d minutes. If you did not sign up, you can ignore this email.\r\n\r\n"+
				"- CODE_LAB\r\n",
			code, expiresInMinutes,
		)
	default:
		subject = "Your CODE_LAB password reset code"
		body = fmt.Sprintf(
			"Hi,\r\n\r\n"+
				"You requested a password reset for your CODE_LAB account.\r\n\r\n"+
				"Your one-time code is:\r\n\r\n"+
				"    %s\r\n\r\n"+
				"This code expires in %d minutes. If you did not request this, you can safely ignore this email.\r\n\r\n"+
				"- CODE_LAB\r\n",
			code, expiresInMinutes,
		)
	}
	return subject, body
}

func generateOTPCode(length int) (string, error) {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", length, n.Int64()), nil
}

func hashOTPCode(email, code string) string {
	return utils.HashToken(strings.ToLower(email) + ":" + strings.TrimSpace(code))
}




