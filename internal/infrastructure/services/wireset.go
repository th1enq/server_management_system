package services

import "github.com/google/wire"

var WireSet = wire.NewSet(
	NewExcelizeService,
	NewBcryptService,
	NewJWTService,
)
