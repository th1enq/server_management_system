package middleware

import "github.com/google/wire"

var WireSet = wire.NewSet(
	NewAuthMiddleware,
	NewMiddlewareManager,

	// Provide default configs
	DefaultMiddlewareManagerConfig,
	DefaultCORSConfig,
	DefaultLoggingConfig,
	DefaultSecurityConfig,
	DefaultRateLimitConfig,
)
