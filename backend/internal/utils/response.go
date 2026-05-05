package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func Success(c *gin.Context, status int, message string, data any) {
	c.JSON(status, APIResponse{
		Error:   false,
		Message: message,
		Data:    data,
	})
}

func Fail(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{
		Error:   true,
		Message: message,
	})
}


func ValidationFail(c *gin.Context, message string, errors map[string]string) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Error:   true,
		Message: message,
		Data: gin.H{
			"errors": errors,
		},
	})
}