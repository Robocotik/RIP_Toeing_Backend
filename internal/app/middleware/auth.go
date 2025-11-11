package middleware

import (
	"backend/internal/app/role"
	"backend/internal/app/service"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const jwtPrefix = "Bearer "

// AuthMiddleware проверяет JWT токен и добавляет пользователя в контекст
func AuthMiddleware(jwtService *service.JWTService, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("=== AUTH MIDDLEWARE STARTED ===")

		tokenString := extractToken(c)
		fmt.Printf("Extracted token string: %s\n", tokenString)

		if tokenString == "" {
			fmt.Println("No token provided")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		// Проверяем, не в черном ли списке токен
		if redisClient != nil {
			isBlacklisted, err := isTokenBlacklisted(c.Request.Context(), redisClient, tokenString)
			if err != nil {
				fmt.Printf("Redis check error: %v\n", err)
			} else if isBlacklisted {
				fmt.Println("Token is in blacklist")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been invalidated"})
				c.Abort()
				return
			}
		}

		// Проверяем валидность токена
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			fmt.Printf("Token validation failed: %v\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid token: %v", err)})
			c.Abort()
			return
		}

		fmt.Printf("Token validated successfully: UserID=%d, Login=%s\n", claims.UserID, claims.Login)

		// Добавляем информацию о пользователе в контекст
		c.Set("userID", claims.UserID)
		c.Set("userUUID", claims.UserUUID)
		c.Set("login", claims.Login)
		c.Set("role", claims.Role)
		c.Set("isModerator", claims.IsModerator)

		fmt.Println("=== AUTH MIDDLEWARE COMPLETED ===")
		c.Next()
	}
}

// isTokenBlacklisted проверяет, находится ли токен в черном списке Redis
func isTokenBlacklisted(ctx context.Context, redisClient *redis.Client, token string) (bool, error) {
	if redisClient == nil {
		return false, nil
	}

	key := "jwt_blacklist:" + token
	val, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil // Токен не найден в черном списке
		}
		return false, err // Ошибка Redis
	}

	return val == "logged_out", nil
}

// extractToken извлекает токен из заголовка Authorization
func extractToken(c *gin.Context) string {
	bearerToken := c.GetHeader("Authorization")
	fmt.Printf("Raw Authorization header: %s\n", bearerToken)

	if bearerToken == "" {
		fmt.Println("Authorization header is empty")
		return ""
	}

	// Если есть префикс "Bearer " - убираем его
	if strings.HasPrefix(bearerToken, "Bearer ") {
		token := bearerToken[len("Bearer "):]
		fmt.Printf("Token after removing Bearer prefix: %s...\n", token[:min(50, len(token))])
		return token
	}

	// Если нет префикса, используем как есть (для Swagger UI)
	fmt.Printf("No Bearer prefix, using token as-is: %s...\n", bearerToken[:min(50, len(bearerToken))])
	return bearerToken
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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