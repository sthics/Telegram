package http

import (
	"net/http"

	"github.com/ambarg/mini-telegram/internal/auth"
	authService "github.com/ambarg/mini-telegram/internal/service/auth"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service *authService.Service
}

func NewAuthHandler(service *authService.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Register(c.Request.Context(), authService.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.setRefreshTokenCookie(c, resp.RefreshToken)
	c.JSON(http.StatusCreated, gin.H{
		"userId":       resp.UserID,
		"accessToken":  resp.AccessToken,
		"refreshToken": resp.RefreshToken,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	h.setRefreshTokenCookie(c, resp.RefreshToken)
	c.JSON(http.StatusOK, gin.H{
		"userId":       resp.UserID,
		"accessToken":  resp.AccessToken,
		"refreshToken": resp.RefreshToken,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refreshToken")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	accessToken, err := h.service.RefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken": accessToken,
	})
}

func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, token string) {
	c.SetCookie("refreshToken", token, int(auth.RefreshTokenLifetime.Seconds()), "/", "", true, true)
}
