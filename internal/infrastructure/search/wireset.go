package search

import "github.com/google/wire"

var WireSet = wire.NewSet(
	LoadElasticSearch,
)
