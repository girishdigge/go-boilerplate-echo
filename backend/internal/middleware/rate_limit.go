package middleware

import "github.com/girishdigge/go-boilerplate-echo/internal/server"

type RateLimitMiddleware struct {
	server *server.Server
}

func NewRateLimitMiddleware(s *server.Server) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		server: s,
	}
}

func (r *RateLimitMiddleware) RecordRateLimitHit(endpoint string) {
	if r == nil || r.server == nil || r.server.LoggerService == nil {
		return
	}

	app := r.server.LoggerService.GetApplication()
	if app == nil {
		return
	}

	app.RecordCustomEvent("RateLimitHit", map[string]any{
		"endpoint": endpoint,
	})
}
