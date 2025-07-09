// @title VCS Server Management System API
// @version 1.0
// @description A comprehensive server management system for monitoring and reporting server status
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	"context"

	_ "github.com/th1enq/server_management_system/docs"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/wiring"
)

const (
	configFilePath = "configs/config.dev.yaml"
)

func main() {
	app, cleanup, err := wiring.InitializeStandardServer(configs.ConfigFilePath(configFilePath))
	if err != nil {
		panic("failed to initialize server: " + err.Error())
	}
	defer cleanup()
	app.Start(context.Background())
}
