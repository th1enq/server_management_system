package tsdb

import "github.com/google/wire"

var WireSet = wire.NewSet(
	NewTSDBClient,
)
