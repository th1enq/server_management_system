package infrastructure

import (
	"github.com/google/wire"
	"github.com/th1enq/server_management_system/internal/infrastructure/cache"
	"github.com/th1enq/server_management_system/internal/infrastructure/database"
	"github.com/th1enq/server_management_system/internal/infrastructure/mq"
	"github.com/th1enq/server_management_system/internal/infrastructure/repository"
	"github.com/th1enq/server_management_system/internal/infrastructure/search"
	"github.com/th1enq/server_management_system/internal/infrastructure/services"
	"github.com/th1enq/server_management_system/internal/infrastructure/tsdb"
)

var WireSet = wire.NewSet(
	cache.WireSet,
	repository.WireSet,
	search.WireSet,
	database.WireSet,
	services.WireSet,
	tsdb.WireSet,
	mq.WireSet,
)
