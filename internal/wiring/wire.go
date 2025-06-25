//go:build wireinject
// +build wireinject

//
//go:generate go run github.com/google/wire/cmd/wire
package wiring

import (
	"github.com/google/wire"
	"github.com/th1enq/server_management_system/internal/configs"
)

var WireSet = wire.NewSet(
	configs.WireSet,
	app.WireSet,
)

func InitializeStandardServer(configFilePath configs.ConfigFilePath) (*app.StandaloneServer, func(), error) {
	wire.Build(WireSet)

	return nil, nil, nil
}
