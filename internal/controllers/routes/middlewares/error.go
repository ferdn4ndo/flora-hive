package middlewares

import (
	"errors"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"

	"flora-hive/internal/domain/models"
	"flora-hive/lib"
)

// ErrorHandlerMiddleware logs and reports handler errors.
type ErrorHandlerMiddleware struct {
	handler lib.RequestHandler
	logger  lib.Logger
}

// NewErrorHandlerMiddleware constructs ErrorHandlerMiddleware.
func NewErrorHandlerMiddleware(handler lib.RequestHandler, logger lib.Logger) ErrorHandlerMiddleware {
	return ErrorHandlerMiddleware{handler: handler, logger: logger}
}

func (m ErrorHandlerMiddleware) sendToSentry(c *gin.Context, e error) {
	if hub := sentrygin.GetHubFromContext(c); hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			if ctx, ok := c.Get("ErrorContext"); ok {
				scope.SetExtra("Context", ctx)
			}
			hub.CaptureException(e)
		})
	}
}

// Setup registers the middleware on /v1.
func (m ErrorHandlerMiddleware) Setup() {
	m.logger.Info("Setting up error handler middleware")
	m.handler.Group.Use(func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		lastError := c.Errors.Last()
		if lastError == nil {
			return
		}
		endTime := time.Now()
		errorContext := map[string]interface{}{
			"startTime":   startTime,
			"endTime":     endTime,
			"latencyTime": endTime.Sub(startTime),
			"reqMethod":   c.Request.Method,
			"reqUri":      c.Request.RequestURI,
			"statusCode":  c.Writer.Status(),
			"clientIP":    c.ClientIP(),
		}
		loggerWithContext := m.logger.With("errorContext", errorContext)
		c.Set("ErrorContext", errorContext)
		for _, ge := range c.Errors {
			err := ge.Err
			if err == nil {
				err = ge
			}
			var logWrapError *models.ErrorWrapped
			if errors.As(err, &logWrapError) {
				loggerWithContext.Error(logWrapError.Unwrap())
				m.sendToSentry(c, logWrapError.Unwrap())
			} else {
				loggerWithContext.Error(err)
				m.sendToSentry(c, err)
			}
		}
		c.JSON(-1, lastError)
	})
}
