package main

import (
	"go.uber.org/fx"

	"flora-hive/internal/controllers"
	"flora-hive/internal/infrastructure"
	"flora-hive/internal/services"
	"flora-hive/lib"
)

// CommonModules is the fx graph shared by serve and migrate commands.
var CommonModules = fx.Options(
	lib.Module,
	controllers.Module,
	services.Module,
	infrastructure.Module,
)
