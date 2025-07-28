package configs

import "github.com/google/wire"

var WireSet = wire.NewSet(
	NewConfig,
	wire.FieldsOf(new(Config), "Server"),
	wire.FieldsOf(new(Config), "Database"),
	wire.FieldsOf(new(Config), "Cache"),
	wire.FieldsOf(new(Config), "Log"),
	wire.FieldsOf(new(Config), "Cron"),
	wire.FieldsOf(new(Config), "JWT"),
	wire.FieldsOf(new(Config), "Elasticsearch"),
	wire.FieldsOf(new(Config), "Email"),
	wire.FieldsOf(new(Config), "TSDB"),
	wire.FieldsOf(new(Config), "MQ"),
)
