package services

import "go.uber.org/fx"

// Module exports domain services.
var Module = fx.Options(
	fx.Provide(NewEnvironmentService),
	fx.Provide(NewDeviceService),
	fx.Provide(NewUserService),
)
