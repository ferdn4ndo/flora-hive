package middlewares

import (
	sentrygin "github.com/getsentry/sentry-go/gin"

	"flora-hive/lib"
)

// SentryMiddleware attaches Sentry to the /v1 group.
type SentryMiddleware struct {
	handler lib.RequestHandler
	logger  lib.Logger
	env     lib.Env
}

// NewSentryMiddleware constructs SentryMiddleware.
func NewSentryMiddleware(handler lib.RequestHandler, logger lib.Logger, env lib.Env) SentryMiddleware {
	return SentryMiddleware{handler: handler, logger: logger, env: env}
}

// Setup registers Sentry.
func (s SentryMiddleware) Setup() {
	s.logger.Info("Setting up sentry middleware")
	s.handler.Group.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
}
