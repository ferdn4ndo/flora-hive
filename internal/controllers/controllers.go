package controllers

import (
	"go.uber.org/fx"

	"flora-hive/internal/controllers/routes"
)

// Module wires HTTP controllers and routes.
var Module = fx.Options(
	routes.Module,
)
