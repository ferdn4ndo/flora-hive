package middlewares

import "go.uber.org/fx"

// Module provides global Gin middleware for /v1.
var Module = fx.Options(
	fx.Provide(NewSentryMiddleware),
	fx.Provide(NewCorsMiddleware),
	fx.Provide(NewErrorHandlerMiddleware),
	fx.Provide(NewMiddlewares),
)

// IMiddleware is applied in order during Setup.
type IMiddleware interface {
	Setup()
}

// Middlewares is an ordered list of middleware.
type Middlewares []IMiddleware

// NewMiddlewares builds the default stack.
func NewMiddlewares(
	sentryMiddleware SentryMiddleware,
	corsMiddleware CorsMiddleware,
	errorHandlerMiddleware ErrorHandlerMiddleware,
) Middlewares {
	return Middlewares{sentryMiddleware, corsMiddleware, errorHandlerMiddleware}
}

// Setup registers all middleware.
func (m Middlewares) Setup() {
	for _, mw := range m {
		mw.Setup()
	}
}
