package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/errs"
	jwtpkg "github.com/mhakimsaputra17/xyz-football-api/pkg/jwt"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/response"
)

// Context keys for storing authenticated admin data.
const (
	ContextKeyAdminID  = "admin_id"
	ContextKeyUsername = "username"
)

// AuthMiddleware returns a GIN middleware that validates JWT access tokens.
// Extracts token from Authorization header, verifies signature and expiration,
// then attaches decoded claims to request context.
func AuthMiddleware(jwtService *jwtpkg.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Abort(c, errs.ErrUnauthorized("Authorization header is required"))
			return
		}

		// Expect "Bearer <token>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Abort(c, errs.ErrUnauthorized("Invalid authorization header format. Use: Bearer <token>"))
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			response.Abort(c, errs.ErrUnauthorized("Access token is required"))
			return
		}

		// Validate and parse the JWT token
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			response.Abort(c, errs.ErrUnauthorized("Invalid or expired access token"))
			return
		}

		// Store admin claims in context for downstream handlers
		c.Set(ContextKeyAdminID, claims.AdminID)
		c.Set(ContextKeyUsername, claims.Username)

		c.Next()
	}
}
