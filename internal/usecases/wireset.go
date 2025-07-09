package usecases

import "github.com/google/wire"

var WireSet = wire.NewSet(
	NewReportService,
	NewServerService,
	NewUserService,
	NewAuthService,
	NewTokenService,
	NewHealthCheckService,
)
