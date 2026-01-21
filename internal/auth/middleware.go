package auth

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

// Middleware returns a Gin middleware for JWT validation
func Middleware(validator *TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{
				"error": "missing authorization header",
			})
			c.Abort()
			return
		}

		// Extract Bearer token
		token, err := ExtractBearerToken(authHeader)
		if err != nil {
			c.JSON(401, gin.H{
				"error": fmt.Sprintf("invalid authorization header: %v", err),
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := validator.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{
				"error": fmt.Sprintf("invalid token: %v", err),
			})
			c.Abort()
			return
		}

		// Store claims in context
		c.Set("user", claims)
		c.Set("username", claims.PreferredUsername)
		c.Set("email", claims.Email)

		log.Printf("✅ Token validated for user: %s", claims.PreferredUsername)
		c.Next()
	}
}

// OptionalMiddleware returns a Gin middleware that doesn't block on missing tokens
func OptionalMiddleware(validator *TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		token, err := ExtractBearerToken(authHeader)
		if err != nil {
			c.Set("user", nil)
			c.Next()
			return
		}

		claims, err := validator.ValidateToken(token)
		if err != nil {
			c.Set("user", nil)
			c.Next()
			return
		}

		c.Set("user", claims)
		c.Set("username", claims.PreferredUsername)
		c.Set("email", claims.Email)
		c.Next()
	}
}

// GetUser extracts the user from context
func GetUser(c *gin.Context) *Claims {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	claims, ok := user.(*Claims)
	if !ok {
		return nil
	}
	return claims
}

// GetUsername extracts the username from context
func GetUsername(c *gin.Context) string {
	username, exists := c.Get("username")
	if !exists {
		return ""
	}
	return username.(string)
}
