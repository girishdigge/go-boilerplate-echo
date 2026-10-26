package middleware

import (
	"errors"
	"net/http"

	"github.com/girishdigge/go-boilerplate-echo/internal/errs"
	"github.com/girishdigge/go-boilerplate-echo/internal/server"
	"github.com/girishdigge/go-boilerplate-echo/internal/sqlerr"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rs/zerolog"
)

type GlobalMiddlewares struct {
	server *server.Server
}

func NewGlobalMiddlewares(s *server.Server) *GlobalMiddlewares {
	return &GlobalMiddlewares{
		server: s,
	}
}

func (global *GlobalMiddlewares) CORS() echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: global.server.Config.Server.CORSAllowedOrigins,
	})
}

func (global *GlobalMiddlewares) RequestLogger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		HandleError: false,
		LogLatency:  true,
		LogHost:     true,
		LogMethod:   true,
		LogURIPath:  true,

		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			statusCode := v.Status

			if v.Error != nil {
				// Preserve committed response status.
				response, unwrapErr := echo.UnwrapResponse(c.Response())
				committed := unwrapErr == nil &&
					response != nil &&
					response.Committed

				if !committed {
					// Handle application-specific HTTP errors.
					if httpErr, ok := errors.AsType[*errs.HTTPError](v.Error); ok && httpErr != nil {
						statusCode = httpErr.Status
					} else if errorStatus := echo.StatusCode(v.Error); errorStatus != 0 {
						// Handle Echo HTTP errors and predefined sentinel errors.
						statusCode = errorStatus
					}
				}
			}

			// Get enhanced logger from context.
			logger := GetLogger(c)

			var e *zerolog.Event

			switch {
			case statusCode >= 500:
				e = logger.Error().Err(v.Error)
			case statusCode >= 400:
				e = logger.Warn()
			default:
				e = logger.Info()
			}

			// Add request ID if available.
			if requestID := GetRequestID(c); requestID != "" {
				e = e.Str("request_id", requestID)
			}

			// Add user context if available.
			if userID := GetUserID(c); userID != "" {
				e = e.Str("user_id", userID)
			}

			e.
				Dur("latency", v.Latency).
				Int("status", statusCode).
				Str("method", v.Method).
				Str("uri", v.URI).
				Str("host", v.Host).
				Str("ip", c.RealIP()).
				Str("user_agent", c.Request().UserAgent()).
				Msg("API")

			return nil
		},
	})
}

func (global *GlobalMiddlewares) Recover() echo.MiddlewareFunc {
	return middleware.Recover()
}

func (global *GlobalMiddlewares) Secure() echo.MiddlewareFunc {
	return middleware.Secure()
}

// GlobalErrorHandler handles centralized application and HTTP errors.
func (global *GlobalMiddlewares) GlobalErrorHandler(c *echo.Context, err error) {
	if c == nil || err == nil {
		return
	}

	// Check whether the response has already been committed.
	// We still process and log the error, but won't overwrite the response.
	response, unwrapErr := echo.UnwrapResponse(c.Response())

	committed := unwrapErr == nil &&
		response != nil &&
		response.Committed

	// Preserve the original error for debugging.
	originalErr := err

	// 1. Check for an existing application HTTP error.
	httpErr, ok := errors.AsType[*errs.HTTPError](err)

	if !ok || httpErr == nil {
		// 2. Check for an Echo HTTP error or predefined status-code error.
		echoStatus := echo.StatusCode(err)

		if echoStatus != 0 {
			// Convert unmatched routes into application-specific 404 errors.
			if echoStatus == http.StatusNotFound {
				err = errs.NewNotFoundError("Route not found", false, nil)
			}
			// Preserve other Echo HTTP errors, such as 405 and 401.
		} else {
			// 3. Convert database and other application errors.
			err = sqlerr.HandleError(err)
		}
	}

	// 4. Extract the final application HTTP error.
	httpErr, _ = errors.AsType[*errs.HTTPError](err)

	var status int
	var code string
	var message string
	var fieldErrors []errs.FieldError
	var action *errs.Action

	switch {
	case httpErr != nil:
		status = httpErr.Status
		code = httpErr.Code
		message = httpErr.Message
		fieldErrors = httpErr.Errors
		action = httpErr.Action

	case echo.StatusCode(err) != 0:
		status = echo.StatusCode(err)
		code = errs.MakeUpperCaseWithUnderscores(
			http.StatusText(status),
		)

		// Echo v5 HTTPError.Message is a string.
		if echoErr, ok := errors.AsType[*echo.HTTPError](err); ok &&
			echoErr != nil && echoErr.Message != "" {
			message = echoErr.Message
		} else {
			message = http.StatusText(status)
		}

	default:
		status = http.StatusInternalServerError
		code = errs.MakeUpperCaseWithUnderscores(
			http.StatusText(http.StatusInternalServerError),
		)
		message = http.StatusText(http.StatusInternalServerError)
	}

	// 5. Log the original error using the enhanced contextual logger.
	logger := GetLogger(c)

	logger.Error().Stack().
		Err(originalErr).
		Int("status", status).
		Str("error_code", code).
		Msg(message)

	// 6. Do not overwrite an already committed response.
	if committed {
		return
	}

	// 7. Send the standardized JSON error response.
	if writeErr := c.JSON(status, errs.HTTPError{
		Code:     code,
		Message:  message,
		Status:   status,
		Override: httpErr != nil && httpErr.Override,
		Errors:   fieldErrors,
		Action:   action,
	}); writeErr != nil {
		logger.Error().
			Err(writeErr).
			Msg("failed to write HTTP error response")
	}
}
