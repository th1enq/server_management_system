package routes

import "github.com/google/wire"

var WireSet = wire.NewSet(
	NewAuthRouter,
	NewServerRouter,
	NewReportRouter,
	NewUserRouter,
	NewJobsRouter,
	NewHandler,
)
