package handlers

import (
	"net/http"

	"cmp/pkg/auth"

	"github.com/gin-gonic/gin"
)

// Login handles user login
func Login(authService auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token, user, err := authService.Login(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user":  user,
		})
	}
}

// Register handles user registration
func Register(authService auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
			Name     string `json:"name" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := authService.Register(c.Request.Context(), req.Email, req.Password, req.Name)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"user": user})
	}
}

// Logout handles user logout
func Logout(authService auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// In a real implementation, you might want to blacklist the token
		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
	}
}

// GetCurrentUser returns the current user information
func GetCurrentUser(authService auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		user, err := authService.GetUser(c.Request.Context(), userID.(string))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": user})
	}
}

// AuthMiddleware validates JWT tokens
func AuthMiddleware(authService auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		userID, err := authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

