package delivery

import (
	"github.com/google/wire"
	"github.com/th1enq/server_management_system/internal/delivery/http"
	"github.com/th1enq/server_management_system/internal/delivery/middleware"
)

var WireSet = wire.NewSet(
	http.WireSet,
	middleware.WireSet,
)
