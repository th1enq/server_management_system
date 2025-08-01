//go:build wireinject
// +build wireinject

//
//go:generate go run github.com/google/wire/cmd/wire
package wiring

import (
	"github.com/google/wire"
	"github.com/th1enq/server_management_system/internal/app"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/delivery"
	"github.com/th1enq/server_management_system/internal/infrastructure"
	"github.com/th1enq/server_management_system/internal/jobs"
	"github.com/th1enq/server_management_system/internal/usecases"
	"github.com/th1enq/server_management_system/internal/utils"
)

var WireSet = wire.NewSet(
	app.WireSet,
	configs.WireSet,
	delivery.WireSet,
	infrastructure.WireSet,
	jobs.WireSet,
	usecases.WireSet,
	utils.WireSet,
)

func InitializeStandardServer(configFilePath configs.ConfigFilePath) (*app.Application, func(), error) {
	wire.Build(WireSet)

	return nil, nil, nil
}
