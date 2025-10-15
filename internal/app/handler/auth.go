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

// RegisterUser godoc
// @Summary Регистрация нового пользователя
// @Description Создает нового пользователя в системе
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body ds.User true "Данные пользователя"
// @Success 201 {object} object "Сообщение об успешной регистрации"
// @Failure 400 {object} object "Неверные данные пользователя"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/auth/register [post]
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

// LoginUser godoc
// @Summary Аутентификация пользователя
// @Description Проверяет учетные данные и возвращает информацию о пользователе
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body object true "Учетные данные" { "login": "user123", "password": "password123" }
// @Success 200 {object} object "Информация о пользователе"
// @Failure 400 {object} object "Неверные данные запроса"
// @Failure 401 {object} object "Неверные учетные данные"
// @Router /api/auth/login [post]
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

// LogoutUser godoc
// @Summary Выход пользователя
// @Description Завершает сеанс пользователя
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} object "Сообщение о выходе"
// @Router /api/auth/logout [post]
func (h *Handler) LogoutUser(ctx *gin.Context) {
	// Для простого токена просто возвращаем успех
	ctx.JSON(http.StatusOK, gin.H{"status": "logged out"})
}

// GetCurrentUser godoc
// @Summary Получить текущего пользователя
// @Description Возвращает информацию о текущем аутентифицированном пользователе
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ds.User "Информация о пользователе"
// @Failure 401 {object} object "Пользователь не аутентифицирован"
// @Failure 404 {object} object "Пользователь не найден"
// @Router /api/auth/me [get]
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

// UpdateCurrentUser godoc
// @Summary Обновить данные текущего пользователя
// @Description Обновляет логин и пароль текущего пользователя
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body AuthRequest true "Новые данные пользователя"
// @Success 200 {object} object "Сообщение об успешном обновлении"
// @Failure 400 {object} object "Неверные данные запроса"
// @Failure 401 {object} object "Пользователь не аутентифицирован"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/auth/me [put]
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