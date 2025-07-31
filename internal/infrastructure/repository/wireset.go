package repository

import "github.com/google/wire"

var WireSet = wire.NewSet(
	NewServerRepository,
	NewUserRepository,
	NewTokenRepository,
)
