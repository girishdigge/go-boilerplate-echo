package middleware

import (
	"context"

	"github.com/girishdigge/go-boilerplate-echo/internal/logger"
	"github.com/girishdigge/go-boilerplate-echo/internal/server"
	"github.com/labstack/echo/v5"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog"
)

const (
	LoggerKey = "logger"
)

type ContextEnhancer struct {
	server *server.Server
}

func NewContextEnhancer(s *server.Server) *ContextEnhancer {
	return &ContextEnhancer{server: s}
}

func (ce *ContextEnhancer) EnhanceContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			//Extract requestID
			requestID := GetRequestID(c)

			//Create enhanced logger with request context
			contextLogger := ce.server.Logger.With().
				Str("request_id", requestID).
				Str("method", c.Request().Method).
				Str("path", c.Path()).
				Str("ip", c.RealIP()).
				Logger()

			//Add trace context if available
			if txn := newrelic.FromContext(c.Request().Context()); txn != nil {
				contextLogger = logger.WithTraceContext(contextLogger, txn)
			}

			//Extract user information from JWT token or session
			if userID := ce.extractUserID(c); userID != "" {
				contextLogger = contextLogger.With().Str("user_id", userID).Logger()
			}

			if userRole := ce.extractUserRole(c); userRole != "" {
				contextLogger = contextLogger.With().Str("user_role", userRole).Logger()
			}

			//Store the enhanced logger in context
			c.Set(LoggerKey, &contextLogger)

			//Create a new context with the logger
			ctx := context.WithValue(c.Request().Context(), LoggerKey, &contextLogger)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func (ce *ContextEnhancer) extractUserID(c *echo.Context) string {
	//Check if user_role was set by auth middleware (clerk)
	userID, err := echo.ContextGet[string](c, UserIDKey)
	if err != nil || userID == "" {
		return ""
	}
	return userID
}
func (ce *ContextEnhancer) extractUserRole(c *echo.Context) string {
	//Check if user_role was set by auth middleware (clerk)
	userRole, err := echo.ContextGet[string](c, UserRoleKey)
	if err != nil || userRole == "" {
		return ""
	}
	return userRole
}

func GetUserID(c *echo.Context) string {
	userID, err := echo.ContextGet[string](c, UserIDKey)
	if err != nil {
		return ""
	}
	return userID
}

func GetUserRole(c *echo.Context) string {
	role, err := echo.ContextGet[string](c, UserRoleKey)
	if err != nil {
		return ""
	}
	return role
}

func GetUserPermissions(c *echo.Context) []string {
	perms, err := echo.ContextGet[[]string](c, PermissionsKey)
	if err != nil {
		return nil
	}
	return perms
}

func GetLogger(c *echo.Context) *zerolog.Logger {
	logger, err := echo.ContextGet[*zerolog.Logger](c, LoggerKey)
	if err == nil && logger != nil {
		return logger
	}

	nopLogger := zerolog.Nop()
	return &nopLogger
}
