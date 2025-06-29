//go:build wireinject
// +build wireinject

//
//go:generate go run github.com/google/wire/cmd/wire
package wiring

import (
	"github.com/google/wire"
	"github.com/th1enq/server_management_system/internal/api"
	"github.com/th1enq/server_management_system/internal/app"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/dataaccess"
	"github.com/th1enq/server_management_system/internal/handler"
	"github.com/th1enq/server_management_system/internal/middleware"
	"github.com/th1enq/server_management_system/internal/repositories"
	"github.com/th1enq/server_management_system/internal/services"
	"github.com/th1enq/server_management_system/internal/utils"
)

var WireSet = wire.NewSet(
	app.WireSet,
	configs.WireSet,
	handler.WireSet,
	dataaccess.WireSet,
	api.WireSet,
	repositories.WireSet,
	services.WireSet,
	middleware.WireSet,
	utils.WireSet,
)

func InitializeStandardServer(configFilePath configs.ConfigFilePath) (*app.StandaloneServer, func(), error) {
	wire.Build(WireSet)

	return nil, nil, nil
}
