package handler

import (
	"backend/internal/app/ds"
	"backend/internal/app/service"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type AuthRequest struct {
	Login    string        `json:"login"`
	Password string        `json:"password"`
	Redis    *redis.Client // Добавляем Redis клиент
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
// @Router /auth/register [post]
func (h *Handler) RegisterUser(ctx *gin.Context) {
	var user ds.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// Устанавливаем роль по умолчанию
	user.IsModerator = false

	if err := h.Repository.RegisterUser(&user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"user_id": user.ID,
	})
}

// LoginUser godoc
// @Summary Аутентификация пользователя
// @Description Проверяет учетные данные и возвращает JWT токен
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body ds.LoginRequest true "Учетные данные"
// @Success 200 {object} ds.LoginResponse "JWT токен и информация о пользователе"
// @Failure 400 {object} object "Неверные данные запроса"
// @Failure 401 {object} object "Неверные учетные данные"
// @Router /auth/login [post]
func (h *Handler) LoginUser(ctx *gin.Context) {
	var creds ds.LoginRequest
	if err := ctx.ShouldBindJSON(&creds); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	user, err := h.Repository.AuthenticateUser(creds.Login, creds.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Используем тот же секрет, что и в middleware
	// Вам нужно передать JWTService из main.go или создать с тем же секретом
	jwtService := service.NewJWTService("your-super-secret-jwt-key-for-flight-api", 24*time.Hour)
	fmt.Printf("Login - JWT Secret: %s\n", "your-super-secret-jwt-key-for-flight-api")

	tokenString, claims, err := jwtService.GenerateToken(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	ctx.JSON(http.StatusOK, ds.LoginResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresAt:   time.Unix(claims.ExpiresAt, 0),
		UserID:      user.ID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
	})
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
// @Router /auth/me [get]
func (h *Handler) GetCurrentUser(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	user, err := h.Repository.GetUserByID(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Не возвращаем пароль
	user.Password = ""
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
// @Router /auth/me [put]
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

// LogoutUser godoc
// @Summary Выход пользователя
// @Description Завершает сеанс пользователя (публичный метод)
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} object "Сообщение о выходе"
// @Router /auth/logout [post]
func (h *Handler) LogoutUser(ctx *gin.Context) {
	fmt.Println("=== LOGOUT REQUEST ===")

	// Пытаемся получить токен из заголовка Authorization для инвалидации
	tokenString := extractTokenFromHeader(ctx)

	if tokenString != "" {
		fmt.Printf("Token found in header, adding to blacklist: %s...\n", tokenString[:min(50, len(tokenString))])

		// Проверяем валидность токена
		claims, err := h.JWTService.ValidateToken(tokenString)
		if err == nil {
			// Токен валидный - добавляем в черный список
			remainingTime := time.Until(time.Unix(claims.ExpiresAt, 0))
			if remainingTime > 0 {
				err = h.addTokenToBlacklist(ctx.Request.Context(), tokenString, remainingTime)
				if err != nil {
					fmt.Printf("Failed to add token to blacklist: %v\n", err)
				} else {
					fmt.Printf("Token added to blacklist, expires in: %v\n", remainingTime)
				}
			}
		} else {
			fmt.Printf("Token validation failed: %v\n", err)
		}
	} else {
		fmt.Println("No token provided for logout - simple logout")
	}

	// Всегда возвращаем успех, даже если токена нет
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "logged out",
		"message": "Logout successful",
	})
}

// extractTokenFromHeader извлекает токен из заголовка Authorization
func extractTokenFromHeader(ctx *gin.Context) string {
	bearerToken := ctx.GetHeader("Authorization")
	if strings.HasPrefix(bearerToken, "Bearer ") {
		return bearerToken[len("Bearer "):]
	}
	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// addTokenToBlacklist добавляет токен в Redis черный список
func (h *Handler) addTokenToBlacklist(ctx context.Context, token string, ttl time.Duration) error {
	if h.Redis == nil {
		return fmt.Errorf("redis client not initialized")
	}

	key := "jwt_blacklist:" + token
	return h.Redis.Set(ctx, key, "logged_out", ttl).Err()
}

// ValidateToken godoc
// @Summary Проверить токен
// @Description Проверяет валидность JWT токена
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object "Информация о токене"
// @Router /auth/validate [get]
func (h *Handler) ValidateToken(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	login, _ := ctx.Get("login")
	isModerator, _ := ctx.Get("isModerator")

	ctx.JSON(http.StatusOK, gin.H{
		"valid":        true,
		"user_id":      userID,
		"login":        login,
		"is_moderator": isModerator,
		"message":      "Token is valid",
	})
}
