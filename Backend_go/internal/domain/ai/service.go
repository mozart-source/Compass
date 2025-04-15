package ai

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Service provides business logic for AI-related operations
type Service interface {
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
	GetUserInteractionHistory(userID uint, domain string, limit, offset int) ([]Interaction, error)
	GetSessionInteractions(sessionID string, limit, offset int) ([]Interaction, error)

	// RAG operations
	AddDocument(input UploadDocumentInput) (*Document, error)
	ProcessDocument(documentID uuid.UUID) error
	GetDocumentByID(id uuid.UUID) (*Document, error)
	ListDocuments(limit, offset int, filters map[string]interface{}) ([]Document, error)
	DeleteDocument(id uuid.UUID) error
	SearchDocuments(query string, domain string, limit int) ([]Document, []Chunk, error)
	GetDocumentContentByID(id uuid.UUID) (string, error)
	GetRagStats(domain string) (map[string]interface{}, error)

	// User context operations
	GetUserContext(userID uint, domain string) (*UserContext, error)
	UpdateUserContext(userID uint, domain string, preferences, history string) (*UserContext, error)
}

type service struct {
	repo   Repository
	logger *logrus.Logger
}

// ServiceConfig contains service configuration options
type ServiceConfig struct {
	Repository Repository
	Logger     *logrus.Logger
}

// NewService creates a new AI service
func NewService(config ServiceConfig) Service {
	return &service{
		repo:   config.Repository,
		logger: config.Logger,
	}
}

// Model operations

func (s *service) CreateModel(input CreateModelInput) (*Model, error) {
	// Check if model already exists
	existingModel, err := s.repo.GetModelByNameAndVersion(input.Name, input.Version)
	if err == nil && existingModel != nil {
		return nil, fmt.Errorf("model with name %s and version %s already exists", input.Name, input.Version)
	}

	// Validate model type
	validTypes := map[string]bool{
		"text-generation":  true,
		"embedding":        true,
		"image-generation": true,
		"speech-to-text":   true,
		"text-to-speech":   true,
		"multimodal":       true,
		"classification":   true,
	}

	if !validTypes[input.Type] {
		return nil, fmt.Errorf("invalid model type: %s", input.Type)
	}

	// Create the model
	s.logger.WithFields(logrus.Fields{
		"name":    input.Name,
		"version": input.Version,
		"type":    input.Type,
	}).Info("Creating new AI model")

	return s.repo.CreateModel(input)
}

func (s *service) GetModelByID(id uuid.UUID) (*Model, error) {
	return s.repo.GetModelByID(id)
}

func (s *service) GetModelByNameAndVersion(name, version string) (*Model, error) {
	return s.repo.GetModelByNameAndVersion(name, version)
}

func (s *service) ListModels(limit, offset int, filters map[string]interface{}) ([]Model, error) {
	// Set default values
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.ListModels(limit, offset, filters)
}

func (s *service) UpdateModel(id uuid.UUID, input UpdateModelInput) (*Model, error) {
	// Get existing model
	model, err := s.repo.GetModelByID(id)
	if err != nil {
		return nil, err
	}

	// Check if update would cause a name/version conflict
	if input.Name != "" && input.Version != "" && (input.Name != model.Name || input.Version != model.Version) {
		existingModel, err := s.repo.GetModelByNameAndVersion(input.Name, input.Version)
		if err == nil && existingModel != nil && existingModel.ID != id {
			return nil, fmt.Errorf("model with name %s and version %s already exists", input.Name, input.Version)
		}
	}

	s.logger.WithFields(logrus.Fields{
		"id":      id,
		"name":    model.Name,
		"version": model.Version,
	}).Info("Updating AI model")

	return s.repo.UpdateModel(id, input)
}

func (s *service) DeleteModel(id uuid.UUID) error {
	// Get existing model
	model, err := s.repo.GetModelByID(id)
	if err != nil {
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"id":      id,
		"name":    model.Name,
		"version": model.Version,
	}).Info("Deleting AI model")

	return s.repo.DeleteModel(id)
}

func (s *service) UpdateModelStats(id uuid.UUID, requestCount int64, latency float64, success bool) error {
	return s.repo.UpdateModelStats(id, requestCount, latency, success)
}

// Interaction operations

func (s *service) LogInteraction(input CreateInteractionInput) (*Interaction, error) {
	// Validate model exists
	model, err := s.repo.GetModelByID(input.ModelID)
	if err != nil {
		return nil, fmt.Errorf("invalid model_id: %v", err)
	}

	// Create interaction log
	s.logger.WithFields(logrus.Fields{
		"model_id":   input.ModelID,
		"model_name": model.Name,
		"user_id":    input.UserID,
		"session_id": input.SessionID,
		"domain":     input.Domain,
	}).Info("Logging AI interaction")

	interaction, err := s.repo.LogInteraction(input)
	if err != nil {
		return nil, err
	}

	// Update model statistics asynchronously (don't wait for this)
	go func() {
		if err := s.UpdateModelStats(input.ModelID, 1, input.Latency, input.Status == "success"); err != nil {
			s.logger.WithError(err).Error("Failed to update model stats")
		}
	}()

	return interaction, nil
}

func (s *service) GetInteractionByID(id uuid.UUID) (*Interaction, error) {
	return s.repo.GetInteractionByID(id)
}

func (s *service) ListInteractions(limit, offset int, filters map[string]interface{}) ([]Interaction, error) {
	// Set default values
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.ListInteractions(limit, offset, filters)
}

func (s *service) AddInteractionFeedback(id uuid.UUID, score float64, text string) error {
	// Validate score range
	if score < 0 || score > 1 {
		return errors.New("feedback score must be between 0 and 1")
	}

	s.logger.WithFields(logrus.Fields{
		"interaction_id": id,
		"score":          score,
	}).Info("Adding feedback to AI interaction")

	return s.repo.AddInteractionFeedback(id, score, text)
}

func (s *service) GetUserInteractionHistory(userID uint, domain string, limit, offset int) ([]Interaction, error) {
	// Set default values
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	filters := map[string]interface{}{
		"user_id": userID,
	}

	if domain != "" {
		filters["domain"] = domain
	}

	return s.repo.ListInteractions(limit, offset, filters)
}

func (s *service) GetSessionInteractions(sessionID string, limit, offset int) ([]Interaction, error) {
	return s.repo.GetInteractionsBySessionID(sessionID, limit, offset)
}

// RAG operations

func (s *service) AddDocument(input UploadDocumentInput) (*Document, error) {
	// Add document to knowledge base
	s.logger.WithFields(logrus.Fields{
		"title":  input.Title,
		"domain": input.Domain,
		"size":   input.FileSize,
	}).Info("Adding document to knowledge base")

	return s.repo.AddDocument(input)
}

func (s *service) ProcessDocument(documentID uuid.UUID) error {
	// Get document
	document, err := s.repo.GetDocumentByID(documentID)
	if err != nil {
		return err
	}

	// Update document status to "processing"
	if err := s.repo.UpdateDocumentStatus(documentID, "processing"); err != nil {
		return err
	}

	// This would typically be a background job in a real system
	// Here we'll do a simple chunking for demonstration

	// Simple chunking by paragraphs
	paragraphs := strings.Split(document.Content, "\n\n")

	s.logger.WithFields(logrus.Fields{
		"document_id": documentID,
		"chunk_count": len(paragraphs),
	}).Info("Processing document into chunks")

	offset := 0
	for i, p := range paragraphs {
		if len(strings.TrimSpace(p)) == 0 {
			continue
		}

		// In a real system, you would generate embeddings here
		// For demo purposes, we'll just store the content
		endOffset := offset + len(p)
		tokenCount := len(strings.Fields(p)) // Simple word count as token estimation

		metadata := fmt.Sprintf(`{"chunk_index": %d, "paragraph": %d}`, i, i)

		_, err := s.repo.AddChunk(
			documentID,
			p,
			"", // placeholder for embedding
			metadata,
			offset,
			endOffset,
			tokenCount,
		)
		if err != nil {
			s.logger.WithError(err).Error("Failed to add chunk")
			// Continue processing other chunks even if one fails
		}

		offset = endOffset + 2 // +2 for the "\n\n" separator
	}

	// Update document status to "processed"
	return s.repo.UpdateDocumentStatus(documentID, "processed")
}

func (s *service) GetDocumentByID(id uuid.UUID) (*Document, error) {
	document, err := s.repo.GetDocumentByID(id)
	if err != nil {
		return nil, err
	}

	// Increment access count
	go s.repo.IncrementDocumentAccessCount(id)

	return document, nil
}

func (s *service) ListDocuments(limit, offset int, filters map[string]interface{}) ([]Document, error) {
	// Set default values
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.ListDocuments(limit, offset, filters)
}

func (s *service) DeleteDocument(id uuid.UUID) error {
	return s.repo.DeleteDocument(id)
}

func (s *service) SearchDocuments(query string, domain string, limit int) ([]Document, []Chunk, error) {
	// Set default values
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	s.logger.WithFields(logrus.Fields{
		"query":  query,
		"domain": domain,
		"limit":  limit,
	}).Info("Searching for documents")

	// Get chunks matching the query
	chunks, err := s.repo.SearchChunks(query, domain, limit)
	if err != nil {
		return nil, nil, err
	}

	// Get the documents for these chunks
	documentIDs := make(map[uuid.UUID]bool)
	for _, chunk := range chunks {
		documentIDs[chunk.DocumentID] = true
	}

	var documents []Document
	for docID := range documentIDs {
		doc, err := s.repo.GetDocumentByID(docID)
		if err != nil {
			s.logger.WithError(err).Error("Failed to get document for search results")
			continue
		}
		documents = append(documents, *doc)
	}

	return documents, chunks, nil
}

func (s *service) GetDocumentContentByID(id uuid.UUID) (string, error) {
	document, err := s.GetDocumentByID(id)
	if err != nil {
		return "", err
	}
	return document.Content, nil
}

func (s *service) GetRagStats(domain string) (map[string]interface{}, error) {
	// Get statistics for the RAG knowledge base
	var filters map[string]interface{}
	if domain != "" {
		filters = map[string]interface{}{"domain": domain}
	}

	documents, err := s.repo.ListDocuments(0, 0, filters)
	if err != nil {
		return nil, err
	}

	// Calculate statistics
	totalDocs := len(documents)
	totalSize := int64(0)
	domainsMap := make(map[string]int)
	statusMap := make(map[string]int)

	for _, doc := range documents {
		totalSize += doc.FileSize
		domainsMap[doc.Domain]++
		statusMap[doc.Status]++
	}

	result := map[string]interface{}{
		"total_documents":  totalDocs,
		"total_size_bytes": totalSize,
		"domains":          domainsMap,
		"status_counts":    statusMap,
		"timestamp":        time.Now().UTC(),
	}

	return result, nil
}

// User context operations

func (s *service) GetUserContext(userID uint, domain string) (*UserContext, error) {
	return s.repo.GetUserContext(userID, domain)
}

func (s *service) UpdateUserContext(userID uint, domain string, preferences, history string) (*UserContext, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"domain":  domain,
	}).Info("Updating user context")

	return s.repo.UpsertUserContext(userID, domain, preferences, history)
}
