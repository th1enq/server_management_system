package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler interface {
	RegisterRoutes() *gin.Engine
}

type handler struct {
	authRouter   AuthRouter
	serverRouter ServerRouter
	reportRouter ReportRouter
	userRouter   UserRouter
	jobsRouter   JobsRouter
}

func NewHandler(
	authRouter AuthRouter,
	serverRouter ServerRouter,
	reportRouter ReportRouter,
	userRouter UserRouter,
	jobsRouter JobsRouter,
) Handler {
	return &handler{
		authRouter:   authRouter,
		serverRouter: serverRouter,
		reportRouter: reportRouter,
		userRouter:   userRouter,
		jobsRouter:   jobsRouter,
	}
}

func (h *handler) RegisterRoutes() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")

	h.authRouter.RegisterRoutes(v1)
	h.serverRouter.RegisterRoutes(v1)
	h.reportRouter.RegisterRoutes(v1)
	h.userRouter.RegisterRoutes(v1)
	h.jobsRouter.RegisterRoutes(v1)

	return router
}
