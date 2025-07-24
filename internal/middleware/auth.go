package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"seta-training/internal/models"
	"seta-training/pkg/auth"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserContextKey      = "user"
	ClaimsContextKey    = "claims"
)

type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// RequireAuth middleware validates JWT token and sets user context
func (a *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := a.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization token required",
			})
			c.Abort()
			return
		}

		claims, err := a.jwtManager.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set claims in context for use in handlers
		c.Set(ClaimsContextKey, claims)
		c.Next()
	}
}

// RequireRole middleware checks if user has required role
func (a *AuthMiddleware) RequireRole(role models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get(ClaimsContextKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*auth.Claims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid token claims",
			})
			c.Abort()
			return
		}

		if userClaims.Role != role {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireManager middleware checks if user is a manager
func (a *AuthMiddleware) RequireManager() gin.HandlerFunc {
	return a.RequireRole(models.RoleManager)
}

// OptionalAuth middleware validates JWT token if present but doesn't require it
func (a *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := a.extractToken(c)
		if token != "" {
			if claims, err := a.jwtManager.ValidateToken(token); err == nil {
				c.Set(ClaimsContextKey, claims)
			}
		}
		c.Next()
	}
}

// extractToken extracts JWT token from Authorization header
func (a *AuthMiddleware) extractToken(c *gin.Context) string {
	authHeader := c.GetHeader(AuthorizationHeader)
	if authHeader == "" {
		return ""
	}

	if !strings.HasPrefix(authHeader, BearerPrefix) {
		return ""
	}

	return strings.TrimPrefix(authHeader, BearerPrefix)
}

// GetCurrentUser returns the current user claims from context
func GetCurrentUser(c *gin.Context) (*auth.Claims, bool) {
	claims, exists := c.Get(ClaimsContextKey)
	if !exists {
		return nil, false
	}

	userClaims, ok := claims.(*auth.Claims)
	return userClaims, ok
}
