package middleware

import (
	"uuid"

	"github.com/labstack/echo/v5"
)

const (
	RequestIDKey = "request_id"
)

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			requestID := c.Request().Header.Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = uuid.New().String()
			}

			c.Set(RequestIDKey, requestID)
			c.Response().Header().Set(echo.HeaderXRequestID, requestID)

			return next(c)
		}
	}
}
func GetRequestID(c *echo.Context) string {

	requestID, err := echo.ContextGet[string](c, RequestIDKey)
	if err != nil {
		return ""
	}
	return requestID
}
