package middlewares

import (
	cors "github.com/rs/cors/wrapper/gin"

	"flora-hive/lib"
)

// CorsMiddleware configures CORS on the /v1 group.
type CorsMiddleware struct {
	handler lib.RequestHandler
	logger  lib.Logger
	env     lib.Env
}

// NewCorsMiddleware constructs CorsMiddleware.
func NewCorsMiddleware(handler lib.RequestHandler, logger lib.Logger, env lib.Env) CorsMiddleware {
	return CorsMiddleware{handler: handler, logger: logger, env: env}
}

// Setup registers CORS.
func (m CorsMiddleware) Setup() {
	m.logger.Info("Setting up cors middleware")
	m.handler.Group.Use(cors.New(cors.Options{
		AllowCredentials: true,
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		Debug:            m.env.IsLocal(),
	}))
}
