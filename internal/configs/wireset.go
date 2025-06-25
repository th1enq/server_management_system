package configs

import "github.com/google/wire"

var WireSet = wire.NewSet(
	NewConfig,
	wire.FieldsOf(new(Config), "server"),
	wire.FieldsOf(new(Config), "database"),
	wire.FieldsOf(new(Config), "cache"),
	wire.FieldsOf(new(Config), "logging"),
	wire.FieldsOf(new(Config), "jwt"),
	wire.FieldsOf(new(Config), "monitoring"),
	wire.FieldsOf(new(Config), "elasticsearch"),
	wire.FieldsOf(new(Config), "email"),
)
