package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	UserID       string   `json:"user_id"`
	Username     string   `json:"username"`
	CompanyID    string   `json:"company_id,omitempty"`
	RoleID       string   `json:"role_id"`
	IsSuperadmin bool     `json:"is_superadmin"`
	Permissions  []string `json:"permissions"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token operations
type JWTManager struct {
	secret []byte
	expiry time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secret string, expiry time.Duration) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		expiry: expiry,
	}
}

// GenerateToken generates a new JWT token with the default expiry
func (j *JWTManager) GenerateToken(userID, username, companyID, roleID string, isSuperadmin bool, permissions []string) (string, error) {
	return j.GenerateTokenWithExpiry(userID, username, companyID, roleID, isSuperadmin, permissions, j.expiry)
}

// GenerateTokenWithExpiry generates a new JWT token with a custom expiry
func (j *JWTManager) GenerateTokenWithExpiry(userID, username, companyID, roleID string, isSuperadmin bool, permissions []string, expiry time.Duration) (string, error) {
	claims := &JWTClaims{
		UserID:       userID,
		Username:     username,
		CompanyID:    companyID,
		RoleID:       roleID,
		IsSuperadmin: isSuperadmin,
		Permissions:  permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// RefreshToken generates a new token with extended expiry
func (j *JWTManager) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	return j.GenerateToken(
		claims.UserID,
		claims.Username,
		claims.CompanyID,
		claims.RoleID,
		claims.IsSuperadmin,
		claims.Permissions,
	)
}

// GetExpiryDuration returns the token expiry duration
func (j *JWTManager) GetExpiryDuration() time.Duration {
	return j.expiry
}
