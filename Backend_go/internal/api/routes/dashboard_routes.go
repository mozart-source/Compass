package routes

import (
	"context"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/handlers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DashboardRoutes struct {
	handler         *handlers.DashboardHandler
	authMiddleware  gin.HandlerFunc
	cacheMiddleware gin.HandlerFunc
	logger          *zap.Logger
}

func NewDashboardRoutes(
	handler *handlers.DashboardHandler,
	authMiddleware gin.HandlerFunc,
	cacheMiddleware gin.HandlerFunc,
	logger *zap.Logger,
) *DashboardRoutes {
	return &DashboardRoutes{
		handler:         handler,
		authMiddleware:  authMiddleware,
		cacheMiddleware: cacheMiddleware,
		logger:          logger,
	}
}

func (r *DashboardRoutes) Register(router *gin.RouterGroup) {
	dashboard := router.Group("/dashboard")
	dashboard.Use(r.authMiddleware)
	{
		dashboard.GET("/metrics", r.cacheMiddleware, r.handler.GetDashboardMetrics)
	}

	// Start the dashboard event listener
	go r.handler.StartDashboardEventListener(context.Background())
}
