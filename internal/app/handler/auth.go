package handler

import (
	"backend/internal/app/ds"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterAuthRoutes(r *gin.Engine) {
	api := r.Group("/api/users")
	{
		api.POST("/register", h.RegisterUser)
		api.POST("/login", h.LoginUser)
	}
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
