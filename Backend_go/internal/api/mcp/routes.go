package mcp

import (
	"github.com/gin-gonic/gin"
)

// Routes contains all the MCP API routes
type Routes struct {
	handler *Handler
}

// NewRoutes creates a new MCP routes instance
func NewRoutes(handler *Handler) *Routes {
	return &Routes{
		handler: handler,
	}
}

// RegisterRoutes registers all MCP API routes
func (r *Routes) RegisterRoutes(rg *gin.RouterGroup) {
	mcpGroup := rg.Group("/mcp")
	{
		// AI model endpoints
		mcpGroup.POST("/ai/model/info", r.handler.GetModelInfo)
		mcpGroup.POST("/ai/model/create", r.handler.CreateModel)
		mcpGroup.POST("/ai/model/stats/update", r.handler.UpdateModelStats)

		// AI interaction endpoints
		mcpGroup.POST("/ai/log/interaction", r.handler.LogInteraction)

		// User context endpoints
		mcpGroup.POST("/user/context", r.handler.GetUserContext)

		// RAG endpoints
		mcpGroup.POST("/rag/stats", r.handler.GetRagStats)
		mcpGroup.POST("/rag/update", r.handler.UpdateRagKnowledge)
		mcpGroup.POST("/rag/knowledge-base/process", r.handler.ProcessKnowledgeBase)
		mcpGroup.POST("/knowledge-base/upload", r.handler.UploadKnowledgeBase)

		// Entity creation endpoint
		mcpGroup.POST("/entity/create", r.handler.CreateEntity)
	}
}
