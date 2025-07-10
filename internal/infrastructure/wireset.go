package infrastructure

import (
	"github.com/google/wire"
	"github.com/th1enq/server_management_system/internal/infrastructure/cache"
	"github.com/th1enq/server_management_system/internal/infrastructure/database"
	"github.com/th1enq/server_management_system/internal/infrastructure/repository"
	"github.com/th1enq/server_management_system/internal/infrastructure/search"
	"github.com/th1enq/server_management_system/internal/infrastructure/services"
)

var WireSet = wire.NewSet(
	cache.WireSet,
	repository.WireSet,
	search.WireSet,
	database.WireSet,
	services.WireSet,
)
