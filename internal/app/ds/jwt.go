package ds

import (
	"backend/internal/app/role"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type JWTClaims struct {
	jwt.StandardClaims
	UserUUID    uuid.UUID `json:"user_uuid"`
	UserID      int       `json:"user_id"`
	Login       string    `json:"login"`
	Role        role.Role `json:"role"`
	IsModerator bool      `json:"is_moderator"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	UserID       int       `json:"user_id"`
	Login        string    `json:"login"`
	IsModerator  bool      `json:"is_moderator"`
}