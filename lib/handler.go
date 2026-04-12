package lib

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
	v1 := engine.Group("/v1")
	return RequestHandler{Gin: engine, Group: v1}
}
