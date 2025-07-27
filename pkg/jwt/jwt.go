package jwt

import (
	"fmt"
	"time"
	"vehicle-showroom-backend/internal/domain/entities"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   int               `json:"user_id"`
	Username string            `json:"username"`
	Role     entities.UserRole `json:"role"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secretKey   string
	expireHours int
}

func NewJWTManager(secretKey string, expireHours int) *JWTManager {
	return &JWTManager{
		secretKey:   secretKey,
		expireHours: expireHours,
	}
}

func (j *JWTManager) GenerateToken(user *entities.User) (string, int64, error) {
	expirationTime := time.Now().Add(time.Duration(j.expireHours) * time.Hour)

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expirationTime.Unix(), nil
}

func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (j *JWTManager) RefreshToken(tokenString string) (string, int64, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", 0, err
	}

	// Check if token is not too old (allow refresh within 7 days of expiration)
	if time.Until(claims.ExpiresAt.Time) < -7*24*time.Hour {
		return "", 0, fmt.Errorf("token too old to refresh")
	}

	// Create new token with same claims but new expiration
	expirationTime := time.Now().Add(time.Duration(j.expireHours) * time.Hour)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	claims.IssuedAt = jwt.NewNumericDate(time.Now())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	newTokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign refreshed token: %w", err)
	}

	return newTokenString, expirationTime.Unix(), nil
}
