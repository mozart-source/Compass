package routes

import (
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/handlers"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// CalendarRoutes handles the setup of calendar-related routes
type CalendarRoutes struct {
	handler   *handlers.CalendarHandler
	jwtSecret string
}

// NewCalendarRoutes creates a new CalendarRoutes instance
func NewCalendarRoutes(handler *handlers.CalendarHandler, jwtSecret string) *CalendarRoutes {
	return &CalendarRoutes{
		handler:   handler,
		jwtSecret: jwtSecret,
	}
}

// RegisterRoutes registers all calendar-related routes
func (cr *CalendarRoutes) RegisterRoutes(router *gin.Engine) {
	// Create a calendar group with authentication middleware
	calendarGroup := router.Group("/api/calendar")
	calendarGroup.Use(middleware.NewAuthMiddleware(cr.jwtSecret))
	//calendarGroup.Use(middleware.OrganizationMiddleware())

	// Event routes
	events := calendarGroup.Group("/events")
	{
		// Collaboration endpoints FIRST
		events.POST("/invite", cr.handler.InviteCollaborator)
		events.POST("/invite/respond", cr.handler.RespondToInvite)
		events.GET("/:id/collaborators", cr.handler.ListCollaborators)
		events.DELETE("/:id/collaborators/:user_id", cr.handler.RemoveCollaborator)

		// Core event operations AFTER
		events.POST("", cr.handler.CreateEvent)
		events.GET("", cr.handler.ListEvents)
		events.GET("/:id", cr.handler.GetEvent)
		events.PUT("/:id", cr.handler.UpdateEvent)
		events.DELETE("/:id", cr.handler.DeleteEvent)

		// Occurrence operations
		events.DELETE("/occurrence", cr.handler.DeleteOccurrence)

		// New ID-based occurrence route
		events.PUT("/occurrences/:id", cr.handler.UpdateOccurrenceById)

		// Reminder operations
		events.POST("/:id/reminders", cr.handler.AddReminder)
	}

	// Shared-with-me endpoint (not in events group, but under /api/calendar/events)
	calendarGroup.GET("/events/shared-with-me", cr.handler.ListEventsSharedWithMe)
}
