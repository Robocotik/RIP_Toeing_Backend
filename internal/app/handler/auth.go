package handler

import (
	"backend/internal/app/ds"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *Handler) RegisterUser(ctx *gin.Context) {
	var user ds.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	if err := h.Repository.RegisterUser(&user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"message": "user registered"})
}

func (h *Handler) LoginUser(ctx *gin.Context) {
	var creds struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := ctx.ShouldBindJSON(&creds); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	user, err := h.Repository.AuthenticateUser(creds.Login, creds.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id":           user.ID,
		"login":        user.Login,
		"is_moderator": user.IsModerator,
	})
}

func (h *Handler) LogoutUser(ctx *gin.Context) {
	// Для простого токена просто возвращаем успех
	ctx.JSON(http.StatusOK, gin.H{"status": "logged out"})
}

// GET /auth/users/me
func (h *Handler) GetCurrentUser(ctx *gin.Context) {
	userIDInterface, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	userID := userIDInterface.(int)

	user, err := h.Repository.GetUserByID(userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// PUT /auth/users/me
func (h *Handler) UpdateCurrentUser(ctx *gin.Context) {
	userIDInterface, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	userID := userIDInterface.(int)

	var req AuthRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	user := ds.User{
		ID:       userID,
		Login:    req.Login,
		Password: req.Password,
	}

	if err := h.Repository.UpdateUser(&user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "updated"})
}