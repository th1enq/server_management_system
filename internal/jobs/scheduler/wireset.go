package scheduler

import "github.com/google/wire"

// WireSet provides dependency injection for scheduler
var WireSet = wire.NewSet(
	NewJobScheduler,
	NewJobManager,
)
