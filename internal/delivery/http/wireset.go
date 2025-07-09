package http

import (
	"github.com/google/wire"
	"github.com/th1enq/server_management_system/internal/delivery/http/controllers"
	"github.com/th1enq/server_management_system/internal/delivery/http/presenters"
)

var WireSet = wire.NewSet(
	controllers.WireSet,
	presenters.WireSet,
	NewController,
	NewServer,
)
