package routes

import (
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/handlers"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// WorkflowRoutes handles the setup of workflow-related routes
type WorkflowRoutes struct {
	handler   *handlers.WorkflowHandler
	jwtSecret string
}

// NewWorkflowRoutes creates a new WorkflowRoutes instance
func NewWorkflowRoutes(handler *handlers.WorkflowHandler, jwtSecret string) *WorkflowRoutes {
	return &WorkflowRoutes{
		handler:   handler,
		jwtSecret: jwtSecret,
	}
}

// RegisterRoutes registers all workflow-related routes
func (wr *WorkflowRoutes) RegisterRoutes(router *gin.Engine) {
	// Create a workflow group with authentication middleware
	workflowGroup := router.Group("/api/workflows")
	workflowGroup.Use(middleware.NewAuthMiddleware(wr.jwtSecret))

	// Core workflow operations
	workflowGroup.POST("", wr.handler.CreateWorkflow)
	workflowGroup.GET("", wr.handler.ListWorkflows)
	workflowGroup.GET("/:id", wr.handler.GetWorkflow)
	workflowGroup.PUT("/:id", wr.handler.UpdateWorkflow)
	workflowGroup.DELETE("/:id", wr.handler.DeleteWorkflow)

	// Workflow steps endpoints
	workflowGroup.POST("/:id/steps", wr.handler.CreateWorkflowStep)
	workflowGroup.GET("/:id/steps", wr.handler.ListWorkflowSteps)
	workflowGroup.GET("/:id/steps/:stepId", wr.handler.GetWorkflowStep)
	workflowGroup.PUT("/:id/steps/:stepId", wr.handler.UpdateWorkflowStep)
	workflowGroup.DELETE("/:id/steps/:stepId", wr.handler.DeleteWorkflowStep)

	// Workflow transitions endpoints
	workflowGroup.POST("/:id/transitions", wr.handler.CreateTransition)
	workflowGroup.GET("/:id/transitions", wr.handler.ListTransitions)
	workflowGroup.GET("/:id/transitions/:transitionId", wr.handler.GetTransition)
	workflowGroup.PUT("/:id/transitions/:transitionId", wr.handler.UpdateTransition)
	workflowGroup.DELETE("/:id/transitions/:transitionId", wr.handler.DeleteTransition)

	// Workflow execution operations
	workflowGroup.POST("/:id/execute", wr.handler.ExecuteWorkflow)
	workflowGroup.POST("/executions/:executionId/cancel", wr.handler.CancelWorkflowExecution)
	workflowGroup.GET("/executions/:executionId", wr.handler.GetWorkflowExecution)
	workflowGroup.GET("/:id/executions", wr.handler.ListWorkflowExecutions)
	workflowGroup.PUT("/step-executions/:executionId", wr.handler.UpdateStepExecution)
	workflowGroup.POST("/step-executions/:executionId/approve", wr.handler.ApproveStepExecution)
	workflowGroup.POST("/step-executions/:executionId/reject", wr.handler.RejectStepExecution)

	// Workflow analysis and optimization
	workflowGroup.GET("/:id/analyze", wr.handler.AnalyzeWorkflow)
	workflowGroup.POST("/:id/optimize", wr.handler.OptimizeWorkflow)
}
