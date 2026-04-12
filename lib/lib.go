package lib

import "go.uber.org/fx"

// Module exports shared infrastructure providers.
var Module = fx.Options(
	fx.Provide(NewRequestHandler),
	fx.Provide(NewEnv),
	fx.Provide(GetLogger),
	fx.Provide(NewDatabase),
)
