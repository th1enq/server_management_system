package http

import (
	"github.com/google/wire"
	"github.com/th1enq/server_management_system/internal/delivery/http/controllers"
	"github.com/th1enq/server_management_system/internal/delivery/http/presenters"
	"github.com/th1enq/server_management_system/internal/delivery/http/routes"
)

var WireSet = wire.NewSet(
	controllers.WireSet,
	presenters.WireSet,
	routes.WireSet,
	NewServer,
)
