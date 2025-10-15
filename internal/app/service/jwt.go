package service

import (
	"backend/internal/app/ds"
	"backend/internal/app/role"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type JWTService struct {
	secretKey     string
	tokenDuration time.Duration
}

func NewJWTService(secretKey string, tokenDuration time.Duration) *JWTService {
	return &JWTService{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

// GenerateToken создает новый JWT токен для пользователя
func (s *JWTService) GenerateToken(user *ds.User) (string, *ds.JWTClaims, error) {
	claims := &ds.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(s.tokenDuration).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "flight-request-api",
		},
		UserUUID:    uuid.New(),
		UserID:      user.ID,
		Login:       user.Login,
		Role:        role.Role(0), // По умолчанию обычный пользователь
		IsModerator: user.IsModerator,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", nil, err
	}

	return tokenString, claims, nil
}

// ValidateToken проверяет валидность JWT токена
func (s *JWTService) ValidateToken(tokenString string) (*ds.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*ds.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}