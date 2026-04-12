package infrastructure

import (
	"go.uber.org/fx"

	"flora-hive/internal/infrastructure/mqtt"
	"flora-hive/internal/infrastructure/repositories"
	"flora-hive/internal/infrastructure/userver"
)

// Module exports infrastructure providers.
var Module = fx.Options(
	repositories.Module,
	fx.Provide(userver.NewClient),
	fx.Provide(mqtt.NewService),
)
