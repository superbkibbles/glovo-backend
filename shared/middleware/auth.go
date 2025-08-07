package middleware

import (
	"net/http"
	"strings"

	"glovo-backend/shared/auth"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens and adds claims to context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("claims", claims)
		c.Next()
	}
}

// RequireRole validates that the user has the required role
func RequireRole(role auth.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No authentication claims found"})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*auth.Claims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid claims format"})
			c.Abort()
			return
		}

		if err := auth.ValidateRole(userClaims, role); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRoles validates that the user has one of the required roles
func RequireRoles(roles []auth.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No authentication claims found"})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*auth.Claims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid claims format"})
			c.Abort()
			return
		}

		if err := auth.ValidateRoles(userClaims, roles); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}
