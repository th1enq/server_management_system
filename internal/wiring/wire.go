//go:build wireinject
// +build wireinject

//
//go:generate go run github.com/google/wire/cmd/wire
package wiring

import (
	"github.com/google/wire"
	"github.com/th1enq/server_management_system/internal/app"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/controller"
	"github.com/th1enq/server_management_system/internal/dataaccess"
	"github.com/th1enq/server_management_system/internal/handler"
	"github.com/th1enq/server_management_system/internal/repositories"
	"github.com/th1enq/server_management_system/internal/services"
	"github.com/th1enq/server_management_system/internal/utils"
)

var WireSet = wire.NewSet(
	app.WireSet,
	configs.WireSet,
	controller.WireSet,
	dataaccess.WireSet,
	handler.WireSet,
	repositories.WireSet,
	services.WireSet,
	utils.WireSet,
)

func InitializeStandardServer(configFilePath configs.ConfigFilePath) (*app.StandaloneServer, func(), error) {
	wire.Build(WireSet)

	return nil, nil, nil
}
