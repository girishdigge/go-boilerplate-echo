package middleware

import (
	"github.com/girishdigge/go-boilerplate-echo/internal/server"
	"github.com/labstack/echo/v5"
	nrecho "github.com/newrelic/go-agent/v3/integrations/nrecho-v5"
	"github.com/newrelic/go-agent/v3/integrations/nrpkgerrors"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type TracingMiddleware struct {
	server *server.Server
	nrApp  *newrelic.Application
}

func NewTracingMiddleware(s *server.Server, nrApp *newrelic.Application) *TracingMiddleware {
	return &TracingMiddleware{
		server: s,
		nrApp:  nrApp,
	}
}

// NewRelicMiddleware returns the New Relic Middleware for Echo
func (tm *TracingMiddleware) NewRelicMiddleware() echo.MiddlewareFunc {
	if tm.nrApp == nil {
		//Return a no-op middleware if New Relic is not initialized
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}
	return nrecho.Middleware(tm.nrApp)
}

// EnhanceTracing adds custom attributes to New Relic transactions
func (tm *TracingMiddleware) EnhanceTracing() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			//Get New Relic transaction from context
			// txn := newrelic.FromContext(c.Request().Context()) [legacy]
			txn := nrecho.FromContext(c)
			if txn == nil {
				return next(c)
			}
			//Add custom attributes
			txn.AddAttribute("service.name", tm.server.Config.Observability.ServiceName)
			txn.AddAttribute("service.environment", tm.server.Config.Observability.Environment)
			txn.AddAttribute("http.real_ip", c.RealIP())
			txn.AddAttribute("http.user_agent", c.Request().UserAgent())

			//Add request ID if available
			if requestID := GetRequestID(c); requestID != "" {
				txn.AddAttribute("request.id", requestID)
			}

			//Add user context if available
			if userID, err := echo.ContextGet[string](c, "user_id"); err == nil {
				txn.AddAttribute("user.id", userID)
			}

			//Execute next handler
			err := next(c)
			//Record error if with enhanced stack traces
			if err != nil {
				txn.NoticeError(nrpkgerrors.Wrap(err))
			}

			// Add response status.
			_, statusCode := echo.ResolveResponseStatus(c.Response(), err)
			txn.AddAttribute("http.status_code", statusCode)

			return err
		}
	}
}
