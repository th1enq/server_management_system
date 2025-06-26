package dataaccess

import (
	"github.com/google/wire"
	"github.com/th1enq/server_management_system/internal/dataaccess/cache"
	"github.com/th1enq/server_management_system/internal/dataaccess/database"
	"github.com/th1enq/server_management_system/internal/dataaccess/elasticsearch"
)

var WireSet = wire.NewSet(
	cache.WireSet,
	database.WireSet,
	elasticsearch.WireSet,
)
