package middleware

import (
	"backend/internal/app/role"
	"backend/internal/app/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const jwtPrefix = "Bearer "

// AuthMiddleware проверяет JWT токен и добавляет пользователя в контекст
func AuthMiddleware(jwtService *service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			logrus.Errorf("Invalid token: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Добавляем информацию о пользователе в контекст
		c.Set("userID", claims.UserID)
		c.Set("userUUID", claims.UserUUID)
		c.Set("login", claims.Login)
		c.Set("role", claims.Role)
		c.Set("isModerator", claims.IsModerator)

		c.Next()
	}
}

// RequireRole проверяет что пользователь имеет необходимую роль
func RequireRole(requiredRole role.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		if userRole.(role.Role) < requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireModerator проверяет что пользователь является модератором
func RequireModerator() gin.HandlerFunc {
	return func(c *gin.Context) {
		isModerator, exists := c.Get("isModerator")
		if !exists || !isModerator.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Moderator access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// extractToken извлекает токен из заголовка Authorization
func extractToken(c *gin.Context) string {
	bearerToken := c.GetHeader("Authorization")
	if strings.HasPrefix(bearerToken, jwtPrefix) {
		return bearerToken[len(jwtPrefix):]
	}
	return ""
}