package mcp

import (
	"net/http"
	"strconv"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/ai"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/user"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Handler handles MCP API requests
type Handler struct {
	aiService   ai.Service
	userService user.Service
	logger      *logrus.Logger
}

// NewHandler creates a new MCP API handler
func NewHandler(aiService ai.Service, userService user.Service, logger *logrus.Logger) *Handler {
	return &Handler{
		aiService:   aiService,
		userService: userService,
		logger:      logger,
	}
}

// GetModelInfo gets AI model information by name and version
func (h *Handler) GetModelInfo(c *gin.Context) {
	var input struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If name and version are provided, look up specific model
	if input.Name != "" && input.Version != "" {
		model, err := h.aiService.GetModelByNameAndVersion(input.Name, input.Version)
		if err != nil {
			// If model not found, create a default response
			h.logger.WithFields(logrus.Fields{
				"name":    input.Name,
				"version": input.Version,
			}).Info("Model not found, returning default info")

			defaultID := uuid.New()
			c.JSON(http.StatusOK, gin.H{
				"model_id":  defaultID,
				"name":      input.Name,
				"version":   input.Version,
				"type":      "text-generation",
				"provider":  "OpenAI",
				"status":    "active",
				"is_active": true,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"model_id":      model.ID,
			"name":          model.Name,
			"version":       model.Version,
			"type":          model.Type,
			"provider":      model.Provider,
			"status":        model.Status,
			"configuration": model.Configuration,
			"is_active":     model.IsActive,
		})
		return
	}

	// Otherwise return default model info
	defaultID := uuid.New()
	c.JSON(http.StatusOK, gin.H{
		"model_id":  defaultID,
		"name":      "gpt-4",
		"version":   "1.0",
		"type":      "text-generation",
		"provider":  "OpenAI",
		"status":    "active",
		"is_active": true,
	})
}

// CreateModel creates a new AI model
func (h *Handler) CreateModel(c *gin.Context) {
	var input ai.CreateModelInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	model, err := h.aiService.CreateModel(input)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create model")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"model_id":      model.ID,
		"name":          model.Name,
		"version":       model.Version,
		"type":          model.Type,
		"provider":      model.Provider,
		"status":        model.Status,
		"configuration": model.Configuration,
		"is_active":     model.IsActive,
	})
}

// UpdateModelStats updates AI model statistics
func (h *Handler) UpdateModelStats(c *gin.Context) {
	var input struct {
		ModelID string  `json:"model_id" binding:"required"`
		Latency float64 `json:"latency" binding:"required"`
		Success bool    `json:"success" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	modelID, err := uuid.Parse(input.ModelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid model ID format"})
		return
	}

	if err := h.aiService.UpdateModelStats(modelID, 1, input.Latency, input.Success); err != nil {
		h.logger.WithError(err).Error("Failed to update model stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"model_id": modelID,
		"updated":  true,
	})
}

// LogInteraction logs an AI interaction
func (h *Handler) LogInteraction(c *gin.Context) {
	var input struct {
		UserID   string                 `json:"user_id"`
		Domain   string                 `json:"domain"`
		Input    string                 `json:"input" binding:"required"`
		Output   string                 `json:"output" binding:"required"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert user_id to uint if provided
	var userID *uint
	if input.UserID != "" {
		uid, err := strconv.ParseUint(input.UserID, 10, 32)
		if err == nil {
			uidUint := uint(uid)
			userID = &uidUint
		}
	}

	// Convert metadata to string
	metadataStr := "{}"
	if len(input.Metadata) > 0 {
		// In a real implementation, you would serialize this to JSON
		metadataStr = `{"intent":"` + input.Metadata["intent"].(string) + `"}`
	}

	// Create default model ID
	defaultModelID := uuid.New()

	// Create interaction log
	interactionInput := ai.CreateInteractionInput{
		ModelID:    defaultModelID, // Default model ID
		UserID:     userID,
		Input:      input.Input,
		Output:     input.Output,
		Domain:     input.Domain,
		Status:     "success",
		Metadata:   metadataStr,
		Intent:     input.Metadata["intent"].(string),
		Confidence: input.Metadata["confidence"].(float64),
		RagUsed:    input.Metadata["rag_used"].(bool),
	}

	interaction, err := h.aiService.LogInteraction(interactionInput)
	if err != nil {
		h.logger.WithError(err).Error("Failed to log interaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logged":         true,
		"interaction_id": interaction.ID,
		"timestamp":      interaction.CreatedAt,
	})
}

// GetUserContext gets user context for a domain
func (h *Handler) GetUserContext(c *gin.Context) {
	var input struct {
		UserID string `json:"user_id" binding:"required"`
		Domain string `json:"domain" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert user_id to uint
	uid, err := strconv.ParseUint(input.UserID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}
	userID := uint(uid)

	// Get user context
	context, err := h.aiService.GetUserContext(userID, input.Domain)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If context not found, return empty context
	if context == nil {
		c.JSON(http.StatusOK, gin.H{
			"user_id":     userID,
			"domain":      input.Domain,
			"preferences": "{}",
			"history":     "[]",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":     context.UserID,
		"domain":      context.Domain,
		"preferences": context.Preferences,
		"history":     context.History,
	})
}

// GetRagStats gets RAG statistics for a domain
func (h *Handler) GetRagStats(c *gin.Context) {
	var input struct {
		Domain string `json:"domain"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stats, err := h.aiService.GetRagStats(input.Domain)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get RAG stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// UpdateRagKnowledge updates RAG knowledge base
func (h *Handler) UpdateRagKnowledge(c *gin.Context) {
	var input struct {
		Domain  string                 `json:"domain" binding:"required"`
		Content map[string]interface{} `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract content from the map
	title, _ := input.Content["title"].(string)
	content, _ := input.Content["content"].(string)
	source, _ := input.Content["source"].(string)

	// Create document
	docInput := ai.UploadDocumentInput{
		Title:      title,
		Content:    content,
		Domain:     input.Domain,
		Source:     source,
		SourceType: "api",
	}

	doc, err := h.aiService.AddDocument(docInput)
	if err != nil {
		h.logger.WithError(err).Error("Failed to add document")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Process document asynchronously
	go func() {
		if err := h.aiService.ProcessDocument(doc.ID); err != nil {
			h.logger.WithError(err).Error("Failed to process document")
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"document_id": doc.ID,
		"domain":      input.Domain,
		"status":      "processing",
		"message":     "Document added and processing started",
	})
}

// UploadKnowledgeBase handles file uploads to the knowledge base
func (h *Handler) UploadKnowledgeBase(c *gin.Context) {
	var input struct {
		Filename string `json:"filename" binding:"required"`
		Content  []byte `json:"content" binding:"required"`
		Domain   string `json:"domain"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determine file type from extension
	fileType := "unknown"
	if len(input.Filename) > 4 {
		ext := input.Filename[len(input.Filename)-4:]
		switch ext {
		case ".pdf":
			fileType = "pdf"
		case ".txt":
			fileType = "text"
		case ".doc", "docx":
			fileType = "document"
		case ".csv":
			fileType = "csv"
		}
	}

	// Create content from binary data
	// In a real implementation, you would parse the file according to its type
	content := string(input.Content[:min(1000, len(input.Content))]) + "..."

	// Create document
	docInput := ai.UploadDocumentInput{
		Title:      input.Filename,
		Content:    content,
		Domain:     input.Domain,
		Source:     "upload",
		SourceType: fileType,
		Filename:   input.Filename,
		FileType:   fileType,
		FileSize:   int64(len(input.Content)),
	}

	doc, err := h.aiService.AddDocument(docInput)
	if err != nil {
		h.logger.WithError(err).Error("Failed to add document")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Process document asynchronously
	go func() {
		if err := h.aiService.ProcessDocument(doc.ID); err != nil {
			h.logger.WithError(err).Error("Failed to process document")
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"status":  "success",
		"message": "File uploaded and processing started",
		"files": []map[string]interface{}{
			{
				"filename":    input.Filename,
				"size":        len(input.Content),
				"document_id": doc.ID,
				"domain":      input.Domain,
				"type":        fileType,
			},
		},
	})
}

// ProcessKnowledgeBase triggers processing of all pending documents
func (h *Handler) ProcessKnowledgeBase(c *gin.Context) {
	var input struct {
		Domain string `json:"domain"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get pending documents
	filters := map[string]interface{}{
		"status": "pending",
	}
	if input.Domain != "" {
		filters["domain"] = input.Domain
	}

	documents, err := h.aiService.ListDocuments(100, 0, filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list pending documents")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Process documents asynchronously
	go func() {
		for _, doc := range documents {
			if err := h.aiService.ProcessDocument(doc.ID); err != nil {
				h.logger.WithError(err).
					WithField("document_id", doc.ID).
					Error("Failed to process document")
			}
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"status":            "processing",
		"pending_documents": len(documents),
		"domain":            input.Domain,
		"message":           "Processing of knowledge base documents started",
	})
}

// CreateEntity creates an entity from a natural language prompt
func (h *Handler) CreateEntity(c *gin.Context) {
	var input struct {
		Prompt string `json:"prompt" binding:"required"`
		Domain string `json:"domain"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// This would typically involve NLP to extract entity information
	// For demo purposes, we'll just return a generic response

	c.JSON(http.StatusOK, gin.H{
		"entity_id":   "123456",
		"response":    "Created entity from prompt: " + input.Prompt,
		"intent":      "create",
		"target":      input.Domain,
		"description": "Entity created from natural language prompt",
		"rag_used":    false,
		"cached":      false,
		"confidence":  0.9,
	})
}

// Helper function for min of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
