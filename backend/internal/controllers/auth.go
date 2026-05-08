package controllers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/ayussh-2/config"
	"github.com/ayussh-2/internal/services"
	"github.com/ayussh-2/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthController struct {
	log         *zap.Logger
	svc         *services.AuthService
	frontendURL string
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

func NewAuthController(log *zap.Logger, svc *services.AuthService, frontendURL string) *AuthController {
	return &AuthController{
		log:         log,
		svc:         svc,
		frontendURL: frontendURL,
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
	}, "", "")
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
	role, _ := c.Get("role")
	utils.Success(c, http.StatusOK, "authenticated user", gin.H{
		"user_id": userID,
		"email":   email,
		"role":    role,
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

var googleOauthConfig = &oauth2.Config{
	ClientID:     config.LoadConfig().GoogleClientID,
	ClientSecret: config.LoadConfig().GoogleClientSecret,
	RedirectURL:  config.LoadConfig().GoogleRedirectURI,
	Scopes: []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	},
	Endpoint: google.Endpoint,
}

func generateRandomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (ac *AuthController) HandleGoogleLogin(c *gin.Context) {
	state := generateRandomState()
	c.SetCookie("oauth_state", state, 3600, "/", "", false, true)

	url := googleOauthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (ac *AuthController) HandleGoogleCallback(c *gin.Context) {
	storedState, _ := c.Cookie("oauth_state")
	if c.Query("state") != storedState {
		utils.Fail(c, http.StatusUnauthorized, "invalid state")
		return
	}

	token, err := googleOauthConfig.Exchange(context.Background(), c.Query("code"))
	if err != nil {
		ac.log.Error("google token exchange failed", zap.Error(err))
		utils.Fail(c, http.StatusInternalServerError, "token exchange failed")
		return
	}

	client := googleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		ac.log.Error("failed to fetch google user info", zap.Error(err))
		utils.Fail(c, http.StatusInternalServerError, "failed to get user info")
		return
	}
	defer resp.Body.Close()

	var userInfo map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		ac.log.Error("failed to decode google user info", zap.Error(err))
		utils.Fail(c, http.StatusInternalServerError, "failed to decode user info")
		return
	}

	email, ok := userInfo["email"].(string)
	if !ok || email == "" {
		utils.Fail(c, http.StatusUnauthorized, "google account email not found")
		return
	}

	name, _ := userInfo["name"].(string)
	result, err := ac.svc.Login(services.LoginInput{}, email, name)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			utils.Fail(c, appErr.Status, appErr.Message)
			return
		}
		utils.Fail(c, http.StatusInternalServerError, "cannot login with google")
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

	userJSON, err := json.Marshal(result.Response.User)
	if err != nil {
		ac.log.Error("failed to encode google user", zap.Error(err))
		utils.Fail(c, http.StatusInternalServerError, "cannot login with google")
		return
	}

	redirectTo := ac.frontendURL + "/auth/google/callback?access_token=" +
		url.QueryEscape(result.Response.AccessToken) +
		"&user=" + url.QueryEscape(base64.RawURLEncoding.EncodeToString(userJSON))
	c.Redirect(http.StatusTemporaryRedirect, redirectTo)
}
