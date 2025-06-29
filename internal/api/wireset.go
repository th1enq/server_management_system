package api

import (
	"github.com/google/wire"
	"github.com/th1enq/server_management_system/internal/api/http"
	"github.com/th1enq/server_management_system/internal/api/jobs"
)

var WireSet = wire.NewSet(
	http.WireSet,
	jobs.WireSet,
)
