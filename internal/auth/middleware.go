package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTMiddleware creates a Gin middleware for JWT authentication
func (s *Service) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "MISSING_TOKEN",
				"message": "authorization token required",
			})
			return
		}

		claims, err := s.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "INVALID_TOKEN",
				"message": "invalid or expired token",
			})
			return
		}

		userID, err := ExtractUserID(claims)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "INVALID_TOKEN",
				"message": "invalid user ID in token",
			})
			return
		}

		// Store user ID in context
		c.Set("uid", userID)
		c.Next()
	}
}

// extractToken extracts the JWT token from the Authorization header
func extractToken(c *gin.Context) string {
	bearerToken := c.GetHeader("Authorization")
	if bearerToken == "" {
		return ""
	}

	// Format: "Bearer <token>"
	parts := strings.SplitN(bearerToken, " ", 2)
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		return parts[1]
	}

	return ""
}

// GetUserID retrieves the user ID from the Gin context
func GetUserID(c *gin.Context) (int64, bool) {
	uid, exists := c.Get("uid")
	if !exists {
		return 0, false
	}

	userID, ok := uid.(int64)
	return userID, ok
}
