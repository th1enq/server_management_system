// Package routes provides HTTP routing configuration for the server management system.
//
// This package implements Clean Architecture principles by separating routing logic
// from business logic. Each domain has its own route file for better organization
// and maintainability.
//
// Structure:
//   - router.go: Main router setup and configuration
//   - auth_routes.go: Authentication and authorization routes
//   - server_routes.go: Server management routes
//   - user_routes.go: User management routes
//   - report_routes.go: Report generation routes
//   - jobs_routes.go: Background jobs monitoring routes
//   - middleware_setup.go: Global middleware configuration
//
// Design Principles:
//   - Separation of Concerns: Each route file handles one domain
//   - Single Responsibility: Routes only handle routing, not business logic
//   - Dependency Injection: Controllers and middleware are injected
//   - Consistent Structure: All route files follow the same pattern
//   - Security First: All routes properly configured with authentication/authorization
//
// Usage:
//
//	config := &routes.RouterConfig{
//	    AuthController: authController,
//	    ServerController: serverController,
//	    // ... other controllers
//	    AuthMiddleware: authMiddleware,
//	}
//	router := routes.NewRouter(config)
//	engine := router.GetEngine()
package routes
