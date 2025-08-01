package outbox

import "github.com/google/wire"

var WireSet = wire.NewSet(
	NewDispatcher,
)
