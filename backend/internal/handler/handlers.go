package handler

import (
	"github.com/girishdigge/go-boilerplate-echo/internal/server"
	"github.com/girishdigge/go-boilerplate-echo/internal/service"
)

type Handlers struct {
	Health  *HealthHandler
	OpenAPI *OpenAPIHandler
}

func NewHandlers(s *server.Server, servcies *service.Services) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(s),
		OpenAPI: NewOpenAPIHandler(s),
	}
}
