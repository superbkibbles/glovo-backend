package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret = []byte("your-secret-key") // In production, use environment variable

type Claims struct {
	UserID string `json:"sub"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleMerchant UserRole = "merchant"
	RoleDriver   UserRole = "driver"
	RoleAdmin    UserRole = "admin"
)

func GenerateToken(userID string, role UserRole) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   string(role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "glovo-backend",
			Subject:   userID,
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func ValidateRole(claims *Claims, requiredRole UserRole) error {
	if claims.Role != string(requiredRole) {
		return errors.New("insufficient permissions")
	}
	return nil
}

func ValidateRoles(claims *Claims, requiredRoles []UserRole) error {
	for _, role := range requiredRoles {
		if claims.Role == string(role) {
			return nil
		}
	}
	return errors.New("insufficient permissions")
}
