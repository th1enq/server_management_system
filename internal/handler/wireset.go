package handler

import (
	"github.com/google/wire"
	"github.com/th1enq/server_management_system/internal/handler/http"
	"github.com/th1enq/server_management_system/internal/handler/jobs"
)

var WireSet = wire.NewSet(
	http.WireSet,
	jobs.WireSet,
)
