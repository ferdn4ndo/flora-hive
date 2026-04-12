package middlewares

import "go.uber.org/fx"

// Module provides Gin middleware for /v1 (CORS is global on the engine in lib.NewRequestHandler).
var Module = fx.Options(
	fx.Provide(NewSentryMiddleware),
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
	errorHandlerMiddleware ErrorHandlerMiddleware,
) Middlewares {
	return Middlewares{sentryMiddleware, errorHandlerMiddleware}
}

// Setup registers all middleware.
func (m Middlewares) Setup() {
	for _, mw := range m {
		mw.Setup()
	}
}
