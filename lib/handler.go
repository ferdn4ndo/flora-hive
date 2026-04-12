package lib

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	corsgin "github.com/rs/cors/wrapper/gin"
)

// RequestHandler holds the Gin engine and the versioned API group (/v1).
type RequestHandler struct {
	Gin   *gin.Engine
	Group *gin.RouterGroup
}

// NewRequestHandler creates Gin with JSON API under /v1.
func NewRequestHandler(logger Logger, env Env) RequestHandler {
	gin.DefaultWriter = logger.GetGinLogger()
	if !env.IsLocal() {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error(recovered)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}))
	// CORS must be on the engine (global middleware), not only on /v1. Browser preflights use OPTIONS and
	// do not match POST-only routes, so group-only CORS never ran and responses had no ACAO header.
	engine.Use(corsgin.New(corsgin.Options{
		AllowCredentials: true,
		AllowOriginFunc:  corsAllowOriginFunc(env),
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		Debug:            env.IsLocal(),
	}))
	v1 := engine.Group("/v1")
	return RequestHandler{Gin: engine, Group: v1}
}

// corsAllowOriginFunc returns a validator for rs/cors. If CORS_ALLOWED_ORIGINS is unset or empty,
// any origin is allowed. Otherwise only listed origins match (comma-separated, trimmed).
func corsAllowOriginFunc(env Env) func(origin string) bool {
	raw := strings.TrimSpace(env.CorsAllowedOriginsRaw)
	if raw == "" {
		return func(string) bool { return true }
	}
	parts := strings.Split(raw, ",")
	allowed := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			allowed = append(allowed, s)
		}
	}
	return func(origin string) bool {
		o := strings.TrimSpace(origin)
		for _, a := range allowed {
			if o == a {
				return true
			}
		}
		return false
	}
}
