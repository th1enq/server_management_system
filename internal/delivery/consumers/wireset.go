package consumers

import "github.com/google/wire"

var WireSet = wire.NewSet(
	NewRoot,
	NewMetricsUpdate,
	NewUpdateStatus,
	NewUptimeCalculator,
)
