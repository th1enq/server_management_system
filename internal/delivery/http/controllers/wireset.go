package controllers

import "github.com/google/wire"

var WireSet = wire.NewSet(
	NewJobsController,
	NewServerController,
	NewUserController,
	NewAuthController,
	NewReportController,
)
