package jobs

import (
	"github.com/google/wire"
	"github.com/th1enq/server_management_system/internal/jobs/scheduler"
	"github.com/th1enq/server_management_system/internal/jobs/tasks"
)

// WireSet provides dependency injection for the entire jobs package
var WireSet = wire.NewSet(
	scheduler.WireSet,
	tasks.WireSet,
)
