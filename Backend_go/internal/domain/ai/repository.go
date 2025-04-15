package ai

import (
	"errors"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/persistence/postgres/connection"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Repository defines the interface for AI-related storage operations
type Repository interface {
	// Model operations
	CreateModel(input CreateModelInput) (*Model, error)
	GetModelByID(id uuid.UUID) (*Model, error)
	GetModelByNameAndVersion(name, version string) (*Model, error)
	ListModels(limit, offset int, filters map[string]interface{}) ([]Model, error)
	UpdateModel(id uuid.UUID, input UpdateModelInput) (*Model, error)
	DeleteModel(id uuid.UUID) error
	UpdateModelStats(id uuid.UUID, requestCount int64, latency float64, success bool) error

	// Interaction operations
	LogInteraction(input CreateInteractionInput) (*Interaction, error)
	GetInteractionByID(id uuid.UUID) (*Interaction, error)
	ListInteractions(limit, offset int, filters map[string]interface{}) ([]Interaction, error)
	AddInteractionFeedback(id uuid.UUID, score float64, text string) error
	GetInteractionsByUserID(userID uint, limit, offset int) ([]Interaction, error)
	GetInteractionsBySessionID(sessionID string, limit, offset int) ([]Interaction, error)

	// RAG document operations
	AddDocument(input UploadDocumentInput) (*Document, error)
	GetDocumentByID(id uuid.UUID) (*Document, error)
	ListDocuments(limit, offset int, filters map[string]interface{}) ([]Document, error)
	DeleteDocument(id uuid.UUID) error
	UpdateDocumentStatus(id uuid.UUID, status string) error
	IncrementDocumentAccessCount(id uuid.UUID) error
	GetDocumentsByDomain(domain string, limit, offset int) ([]Document, error)

	// RAG chunk operations
	AddChunk(documentID uuid.UUID, content string, embedding string, metadata string, startOffset, endOffset, tokenCount int) (*Chunk, error)
	GetChunksByDocumentID(documentID uuid.UUID) ([]Chunk, error)
	SearchChunks(query string, domain string, limit int) ([]Chunk, error)
	IncrementChunkRetrievalCount(id uuid.UUID) error

	// User context operations
	GetUserContext(userID uint, domain string) (*UserContext, error)
	UpsertUserContext(userID uint, domain string, preferences, history string) (*UserContext, error)
}

type repository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewRepository creates a new repository for AI operations
func NewRepository(db *connection.Database, logger *logrus.Logger) Repository {
	return &repository{
		db:     db.DB,
		logger: logger,
	}
}

// Model operations

func (r *repository) CreateModel(input CreateModelInput) (*Model, error) {
	model := &Model{
		Name:          input.Name,
		Version:       input.Version,
		Type:          input.Type,
		Provider:      input.Provider,
		Status:        input.Status,
		Configuration: input.Configuration,
		IsActive:      true,
	}

	if err := r.db.Create(model).Error; err != nil {
		r.logger.WithError(err).Error("Failed to create AI model")
		return nil, err
	}

	return model, nil
}

func (r *repository) GetModelByID(id uuid.UUID) (*Model, error) {
	var model Model
	if err := r.db.First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.WithError(err).Error("Failed to get AI model by ID")
		return nil, err
	}
	return &model, nil
}

func (r *repository) GetModelByNameAndVersion(name, version string) (*Model, error) {
	var model Model
	if err := r.db.Where("name = ? AND version = ?", name, version).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.WithError(err).Error("Failed to get AI model by name and version")
		return nil, err
	}
	return &model, nil
}

func (r *repository) ListModels(limit, offset int, filters map[string]interface{}) ([]Model, error) {
	var models []Model
	query := r.db.Model(&Model{})

	// Apply filters
	for key, value := range filters {
		query = query.Where(key, value)
	}

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		r.logger.WithError(err).Error("Failed to list AI models")
		return nil, err
	}

	return models, nil
}

func (r *repository) UpdateModel(id uuid.UUID, input UpdateModelInput) (*Model, error) {
	model, err := r.GetModelByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Name != "" {
		model.Name = input.Name
	}
	if input.Version != "" {
		model.Version = input.Version
	}
	if input.Type != "" {
		model.Type = input.Type
	}
	if input.Provider != "" {
		model.Provider = input.Provider
	}
	if input.Status != "" {
		model.Status = input.Status
	}
	if input.Configuration != "" {
		model.Configuration = input.Configuration
	}
	if input.IsActive != nil {
		model.IsActive = *input.IsActive
	}

	if err := r.db.Save(model).Error; err != nil {
		r.logger.WithError(err).Error("Failed to update AI model")
		return nil, err
	}

	return model, nil
}

func (r *repository) DeleteModel(id uuid.UUID) error {
	if err := r.db.Delete(&Model{}, id).Error; err != nil {
		r.logger.WithError(err).Error("Failed to delete AI model")
		return err
	}
	return nil
}

func (r *repository) UpdateModelStats(id uuid.UUID, requestCount int64, latency float64, success bool) error {
	model, err := r.GetModelByID(id)
	if err != nil {
		return err
	}

	// Update statistics
	newCount := model.RequestCount + requestCount
	newAvgLatency := ((model.AvgLatency * float64(model.RequestCount)) + latency) / float64(newCount)

	// Update error rate if needed
	var newErrorRate float64
	if !success {
		newErrorRate = ((model.ErrorRate * float64(model.RequestCount)) + 1) / float64(newCount)
	} else {
		newErrorRate = (model.ErrorRate * float64(model.RequestCount)) / float64(newCount)
	}

	now := time.Now()
	model.RequestCount = newCount
	model.AvgLatency = newAvgLatency
	model.ErrorRate = newErrorRate
	model.LastUsed = &now

	if err := r.db.Save(model).Error; err != nil {
		r.logger.WithError(err).Error("Failed to update AI model stats")
		return err
	}

	return nil
}

// Interaction operations

func (r *repository) LogInteraction(input CreateInteractionInput) (*Interaction, error) {
	interaction := &Interaction{
		ModelID:      input.ModelID,
		UserID:       input.UserID,
		SessionID:    input.SessionID,
		Input:        input.Input,
		Output:       input.Output,
		Latency:      input.Latency,
		TokenCount:   input.TokenCount,
		InputTokens:  input.InputTokens,
		OutputTokens: input.OutputTokens,
		Domain:       input.Domain,
		Status:       input.Status,
		Error:        input.Error,
		Metadata:     input.Metadata,
		RagUsed:      input.RagUsed,
		Cached:       input.Cached,
		Intent:       input.Intent,
		Confidence:   input.Confidence,
	}

	if err := r.db.Create(interaction).Error; err != nil {
		r.logger.WithError(err).Error("Failed to log AI interaction")
		return nil, err
	}

	return interaction, nil
}

func (r *repository) GetInteractionByID(id uuid.UUID) (*Interaction, error) {
	var interaction Interaction
	if err := r.db.First(&interaction, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.WithError(err).Error("Failed to get AI interaction by ID")
		return nil, err
	}
	return &interaction, nil
}

func (r *repository) ListInteractions(limit, offset int, filters map[string]interface{}) ([]Interaction, error) {
	var interactions []Interaction
	query := r.db.Model(&Interaction{})

	// Apply filters
	for key, value := range filters {
		query = query.Where(key, value)
	}

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&interactions).Error; err != nil {
		r.logger.WithError(err).Error("Failed to list AI interactions")
		return nil, err
	}

	return interactions, nil
}

func (r *repository) AddInteractionFeedback(id uuid.UUID, score float64, text string) error {
	interaction, err := r.GetInteractionByID(id)
	if err != nil {
		return err
	}

	interaction.FeedbackScore = &score
	interaction.FeedbackText = text

	if err := r.db.Save(interaction).Error; err != nil {
		r.logger.WithError(err).Error("Failed to add feedback to AI interaction")
		return err
	}

	return nil
}

func (r *repository) GetInteractionsByUserID(userID uint, limit, offset int) ([]Interaction, error) {
	var interactions []Interaction
	query := r.db.Where("user_id = ?", userID)

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&interactions).Error; err != nil {
		r.logger.WithError(err).Error("Failed to get AI interactions by user ID")
		return nil, err
	}

	return interactions, nil
}

func (r *repository) GetInteractionsBySessionID(sessionID string, limit, offset int) ([]Interaction, error) {
	var interactions []Interaction
	query := r.db.Where("session_id = ?", sessionID)

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&interactions).Error; err != nil {
		r.logger.WithError(err).Error("Failed to get AI interactions by session ID")
		return nil, err
	}

	return interactions, nil
}

// RAG document operations

func (r *repository) AddDocument(input UploadDocumentInput) (*Document, error) {
	document := &Document{
		Title:      input.Title,
		Content:    input.Content,
		Domain:     input.Domain,
		Source:     input.Source,
		SourceType: input.SourceType,
		UploadedBy: input.UploadedBy,
		Filename:   input.Filename,
		FileType:   input.FileType,
		FileSize:   input.FileSize,
		Status:     "pending", // Default status
		Tags:       input.Tags,
	}

	if err := r.db.Create(document).Error; err != nil {
		r.logger.WithError(err).Error("Failed to add document to knowledge base")
		return nil, err
	}

	return document, nil
}

func (r *repository) GetDocumentByID(id uuid.UUID) (*Document, error) {
	var document Document
	if err := r.db.First(&document, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.WithError(err).Error("Failed to get document by ID")
		return nil, err
	}
	return &document, nil
}

func (r *repository) ListDocuments(limit, offset int, filters map[string]interface{}) ([]Document, error) {
	var documents []Document
	query := r.db.Model(&Document{})

	// Apply filters
	for key, value := range filters {
		query = query.Where(key, value)
	}

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&documents).Error; err != nil {
		r.logger.WithError(err).Error("Failed to list documents")
		return nil, err
	}

	return documents, nil
}

func (r *repository) DeleteDocument(id uuid.UUID) error {
	if err := r.db.Delete(&Document{}, id).Error; err != nil {
		r.logger.WithError(err).Error("Failed to delete document")
		return err
	}
	return nil
}

func (r *repository) UpdateDocumentStatus(id uuid.UUID, status string) error {
	document, err := r.GetDocumentByID(id)
	if err != nil {
		return err
	}

	document.Status = status
	now := time.Now()
	document.LastProcessed = &now

	if err := r.db.Save(document).Error; err != nil {
		r.logger.WithError(err).Error("Failed to update document status")
		return err
	}

	return nil
}

func (r *repository) IncrementDocumentAccessCount(id uuid.UUID) error {
	return r.db.Model(&Document{}).Where("id = ?", id).
		UpdateColumn("access_count", gorm.Expr("access_count + ?", 1)).Error
}

func (r *repository) GetDocumentsByDomain(domain string, limit, offset int) ([]Document, error) {
	var documents []Document
	query := r.db.Where("domain = ?", domain)

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&documents).Error; err != nil {
		r.logger.WithError(err).Error("Failed to get documents by domain")
		return nil, err
	}

	return documents, nil
}

// RAG chunk operations

func (r *repository) AddChunk(documentID uuid.UUID, content string, embedding string, metadata string, startOffset, endOffset, tokenCount int) (*Chunk, error) {
	chunk := &Chunk{
		DocumentID:  documentID,
		Content:     content,
		Embedding:   embedding,
		StartOffset: startOffset,
		EndOffset:   endOffset,
		TokenCount:  tokenCount,
		Metadata:    metadata,
	}

	if err := r.db.Create(chunk).Error; err != nil {
		r.logger.WithError(err).Error("Failed to add chunk")
		return nil, err
	}

	return chunk, nil
}

func (r *repository) GetChunksByDocumentID(documentID uuid.UUID) ([]Chunk, error) {
	var chunks []Chunk
	if err := r.db.Where("document_id = ?", documentID).Find(&chunks).Error; err != nil {
		r.logger.WithError(err).Error("Failed to get chunks by document ID")
		return nil, err
	}
	return chunks, nil
}

func (r *repository) SearchChunks(query string, domain string, limit int) ([]Chunk, error) {
	// This would use vector search in a real implementation
	// For now, we'll do a simple text search
	var chunks []Chunk
	textQuery := r.db.Table("rag_chunks").
		Joins("JOIN rag_documents ON rag_chunks.document_id = rag_documents.id").
		Where("rag_documents.domain = ?", domain).
		Where("rag_chunks.content ILIKE ?", "%"+query+"%")

	if limit > 0 {
		textQuery = textQuery.Limit(limit)
	}

	if err := textQuery.Find(&chunks).Error; err != nil {
		r.logger.WithError(err).Error("Failed to search chunks")
		return nil, err
	}

	// Update retrieval counts for the found chunks
	for _, chunk := range chunks {
		go r.IncrementChunkRetrievalCount(chunk.ID) // Non-blocking update
	}

	return chunks, nil
}

func (r *repository) IncrementChunkRetrievalCount(id uuid.UUID) error {
	return r.db.Model(&Chunk{}).Where("id = ?", id).
		UpdateColumn("retrieval_count", gorm.Expr("retrieval_count + ?", 1)).Error
}

// User context operations

func (r *repository) GetUserContext(userID uint, domain string) (*UserContext, error) {
	var context UserContext
	if err := r.db.Where("user_id = ? AND domain = ?", userID, domain).First(&context).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No context found
		}
		r.logger.WithError(err).Error("Failed to get user context")
		return nil, err
	}
	return &context, nil
}

func (r *repository) UpsertUserContext(userID uint, domain string, preferences, history string) (*UserContext, error) {
	// Check if context already exists
	context, err := r.GetUserContext(userID, domain)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if context == nil {
		// Create new context
		context = &UserContext{
			UserID:         userID,
			Domain:         domain,
			Preferences:    preferences,
			History:        history,
			LastInteracted: now,
		}
		if err := r.db.Create(context).Error; err != nil {
			r.logger.WithError(err).Error("Failed to create user context")
			return nil, err
		}
	} else {
		// Update existing context
		if preferences != "" {
			context.Preferences = preferences
		}
		if history != "" {
			context.History = history
		}
		context.LastInteracted = now
		if err := r.db.Save(context).Error; err != nil {
			r.logger.WithError(err).Error("Failed to update user context")
			return nil, err
		}
	}

	return context, nil
}
