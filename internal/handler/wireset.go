package handler

import "github.com/google/wire"

var WireSet = wire.NewSet(
	NewReportHandler,
	NewServerHandler,
	NewAuthHandler,
	NewUserHandler,
)
