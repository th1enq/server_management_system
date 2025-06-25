package dataacess

import (
	"github.com/google/wire"
	"github.com/th1enq/server_management_system/internal/dataacess/cache"
	"github.com/th1enq/server_management_system/internal/dataacess/database"
	"github.com/th1enq/server_management_system/internal/dataacess/elasticsearch"
)

var WireSet = wire.NewSet(
	cache.WireSet,
	database.WireSet,
	elasticsearch.WireSet,
)
