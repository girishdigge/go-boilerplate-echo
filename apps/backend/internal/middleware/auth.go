package middleware

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/girishdigge/go-boilerplate-echo/internal/errs"
	"github.com/girishdigge/go-boilerplate-echo/internal/server"
	"github.com/labstack/echo/v5"
)

const (
	UserIDKey      = "user_id"
	UserRoleKey    = "user_role"
	PermissionsKey = "permissions"
)

type AuthMiddleware struct {
	server *server.Server
}

func NewAuthMiddleware(s *server.Server) *AuthMiddleware {
	return &AuthMiddleware{
		server: s,
	}
}

func (auth *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	//  2: Extract the failure handler into a variable for readability.
	failureHandler := clerkhttp.AuthorizationFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)

		response := map[string]string{
			"code":     "UNAUTHORIZED",
			"message":  "Unauthorized",
			"override": "false",
			"status":   "401",
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			auth.server.Logger.Error().Err(err).Str("function", "RequireAuth").Dur(
				"duration", time.Since(start)).Msg("failed to write JSON response")
		}
	}))

	//3: Un-nest the `WrapMiddleware` call so the Echo v5 handler is clearly separated.
	clerkEchoMiddleware := echo.WrapMiddleware(clerkhttp.WithHeaderAuthorization(failureHandler))

	return clerkEchoMiddleware(func(c *echo.Context) error {
		start := time.Now()
		claims, ok := clerk.SessionClaimsFromContext(c.Request().Context())

		if !ok {
			auth.server.Logger.Error().
				Str("function", "RequireAuth").
				Str("request_id", GetRequestID(c)).
				Dur("duration", time.Since(start)).
				Msg("could not get session claims from context")
			return errs.NewUnauthorizedError("Unauthorized", false)
		}

		c.Set(UserIDKey, claims.Subject)
		c.Set(UserRoleKey, claims.ActiveOrganizationRole)
		c.Set(PermissionsKey, claims.ActiveOrganizationPermissions)

		auth.server.Logger.Info().
			Str("function", "RequireAuth").
			Str("user_id", claims.Subject).
			Str("request_id", GetRequestID(c)).
			Dur("duration", time.Since(start)).
			Msg("user authenticated successfully")

		return next(c)
	})
}
