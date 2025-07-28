package tasks

import "github.com/google/wire"

var WireSet = wire.NewSet(
	NewDailyReportTask,
	NewUpdateStatusTask,
)
