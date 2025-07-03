package database

import "github.com/google/wire"

var WireSet = wire.NewSet(
	LoadPostgres,
	LoadRedis,
	LoadElasticSearch,
)
