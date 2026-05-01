package main

import (
	fxmodule "telemetry-collector/internal/infrastructure/fx"

	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fxmodule.Module(),
	)
	app.Run()
}
