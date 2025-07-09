package usecases

import "github.com/google/wire"

var WireSet = wire.NewSet(
	NewServerUseCase,
	NewUserUseCase,
	NewTokenUseCase,
	NewHealthCheckUseCase,
	NewReportUseCase,
	NewAuthUseCase,
)
