package middleware

import (
	"net/http"
	"skyclust/internal/domain"
	"skyclust/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RBACMiddleware creates a middleware for role-based access control
func RBACMiddleware(rbacService domain.RBACService, requiredPermissions []domain.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (set by AuthMiddleware)
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "User not authenticated",
				"code":    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		// Extract user ID
		var userID uuid.UUID
		if domainUser, ok := user.(*domain.User); ok {
			userID = domainUser.ID
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid user data",
				"code":    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		// Check permissions
		hasPermission, err := rbacService.CheckAnyPermission(userID, requiredPermissions)
		if err != nil {
			logger.Errorf("Failed to check permissions for user %s: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Permission check failed",
				"code":    "INTERNAL_ERROR",
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Insufficient permissions",
				"code":    "FORBIDDEN",
			})
			c.Abort()
			return
		}

		// Set user ID in context for use in handlers
		c.Set("user_id", userID.String())
		c.Next()
	}
}

// AdminMiddleware creates a middleware that requires admin role
func AdminMiddleware(rbacService domain.RBACService) gin.HandlerFunc {
	return RBACMiddleware(rbacService, []domain.Permission{
		domain.SystemManage,
		domain.UserManage,
		domain.AuditManage,
	})
}

// UserManagementMiddleware creates a middleware for user management operations
func UserManagementMiddleware(rbacService domain.RBACService) gin.HandlerFunc {
	return RBACMiddleware(rbacService, []domain.Permission{
		domain.UserManage,
		domain.UserUpdate,
		domain.UserDelete,
	})
}

// SystemManagementMiddleware creates a middleware for system management operations
func SystemManagementMiddleware(rbacService domain.RBACService) gin.HandlerFunc {
	return RBACMiddleware(rbacService, []domain.Permission{
		domain.SystemManage,
		domain.SystemUpdate,
	})
}

// AuditManagementMiddleware creates a middleware for audit log operations
func AuditManagementMiddleware(rbacService domain.RBACService) gin.HandlerFunc {
	return RBACMiddleware(rbacService, []domain.Permission{
		domain.AuditManage,
		domain.AuditRead,
		domain.AuditExport,
	})
}
