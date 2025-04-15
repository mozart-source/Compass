package ai

import (
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/base"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Model represents an AI model configuration in the system
type Model struct {
	base.Model
	Name          string        `json:"name" gorm:"uniqueIndex:idx_model_name_version"`
	Version       string        `json:"version" gorm:"uniqueIndex:idx_model_name_version"`
	Type          string        `json:"type" gorm:"index:idx_model_type"`
	Provider      string        `json:"provider" gorm:"index:idx_model_provider"`
	Status        string        `json:"status" gorm:"index:idx_model_status"`
	Configuration string        `json:"configuration" gorm:"type:jsonb"`
	RequestCount  int64         `json:"request_count" gorm:"default:0"`
	AvgLatency    float64       `json:"avg_latency" gorm:"default:0"`
	ErrorRate     float64       `json:"error_rate" gorm:"default:0"`
	LastUsed      *time.Time    `json:"last_used"`
	Interactions  []Interaction `json:"interactions,omitempty" gorm:"foreignKey:ModelID"`
	IsActive      bool          `json:"is_active" gorm:"default:true"`
}

// ModelConfig represents configuration options for a model
type ModelConfig struct {
	Temperature      float64 `json:"temperature"`
	MaxTokens        int     `json:"max_tokens"`
	TopP             float64 `json:"top_p"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
	PresencePenalty  float64 `json:"presence_penalty"`
	UseCache         bool    `json:"use_cache"`
}

// Interaction represents an AI interaction log
type Interaction struct {
	base.Model
	ModelID       uuid.UUID `json:"model_id" gorm:"index:idx_interaction_model;type:uuid"`
	UserID        *uint     `json:"user_id" gorm:"index:idx_interaction_user"`
	SessionID     string    `json:"session_id" gorm:"index:idx_interaction_session"`
	Input         string    `json:"input" gorm:"type:text"`
	Output        string    `json:"output" gorm:"type:text"`
	Latency       float64   `json:"latency"`
	TokenCount    int       `json:"token_count"`
	InputTokens   int       `json:"input_tokens"`
	OutputTokens  int       `json:"output_tokens"`
	Domain        string    `json:"domain" gorm:"index:idx_interaction_domain"`
	Status        string    `json:"status" gorm:"index:idx_interaction_status"`
	Error         string    `json:"error" gorm:"type:text"`
	Metadata      string    `json:"metadata" gorm:"type:jsonb"`
	FeedbackScore *float64  `json:"feedback_score"`
	FeedbackText  string    `json:"feedback_text" gorm:"type:text"`
	RagUsed       bool      `json:"rag_used" gorm:"default:false"`
	Cached        bool      `json:"cached" gorm:"default:false"`
	Intent        string    `json:"intent" gorm:"index:idx_interaction_intent"`
	Confidence    float64   `json:"confidence"`
}

// Document represents a document in the RAG knowledge base
type Document struct {
	base.Model
	Title         string     `json:"title" gorm:"index:idx_document_title"`
	Content       string     `json:"content" gorm:"type:text"`
	Domain        string     `json:"domain" gorm:"index:idx_document_domain"`
	Source        string     `json:"source" gorm:"index:idx_document_source"`
	SourceType    string     `json:"source_type"`
	UploadedBy    *uint      `json:"uploaded_by" gorm:"index:idx_document_uploader"`
	Filename      string     `json:"filename"`
	FileType      string     `json:"file_type"`
	FileSize      int64      `json:"file_size"`
	Hash          string     `json:"hash" gorm:"uniqueIndex:idx_document_hash"`
	Status        string     `json:"status" gorm:"index:idx_document_status"`
	Chunks        []Chunk    `json:"chunks,omitempty" gorm:"foreignKey:DocumentID"`
	LastProcessed *time.Time `json:"last_processed"`
	AccessCount   int64      `json:"access_count" gorm:"default:0"`
	Tags          string     `json:"tags" gorm:"type:jsonb"`
}

// Chunk represents a document chunk with embeddings
type Chunk struct {
	base.Model
	DocumentID     uuid.UUID `json:"document_id" gorm:"index:idx_chunk_document;type:uuid"`
	Content        string    `json:"content" gorm:"type:text"`
	Embedding      string    `json:"embedding" gorm:"type:text"`
	StartOffset    int       `json:"start_offset"`
	EndOffset      int       `json:"end_offset"`
	TokenCount     int       `json:"token_count"`
	Metadata       string    `json:"metadata" gorm:"type:jsonb"`
	RetrievalCount int64     `json:"retrieval_count" gorm:"default:0"`
}

// UserContext represents personalized context for a user
type UserContext struct {
	base.Model
	UserID         uint      `json:"user_id" gorm:"uniqueIndex:idx_user_context_domain"`
	Domain         string    `json:"domain" gorm:"uniqueIndex:idx_user_context_domain"`
	Preferences    string    `json:"preferences" gorm:"type:jsonb"`
	History        string    `json:"history" gorm:"type:jsonb"`
	LastInteracted time.Time `json:"last_interacted"`
}

// CreateModelInput is used to create a new model
type CreateModelInput struct {
	Name          string `json:"name" binding:"required"`
	Version       string `json:"version" binding:"required"`
	Type          string `json:"type" binding:"required"`
	Provider      string `json:"provider" binding:"required"`
	Status        string `json:"status" binding:"required"`
	Configuration string `json:"configuration"`
}

// UpdateModelInput is used to update an existing model
type UpdateModelInput struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	Type          string `json:"type"`
	Provider      string `json:"provider"`
	Status        string `json:"status"`
	Configuration string `json:"configuration"`
	IsActive      *bool  `json:"is_active"`
}

// CreateInteractionInput is used to log a new interaction
type CreateInteractionInput struct {
	ModelID      uuid.UUID `json:"model_id" binding:"required"`
	UserID       *uint     `json:"user_id"`
	SessionID    string    `json:"session_id"`
	Input        string    `json:"input" binding:"required"`
	Output       string    `json:"output" binding:"required"`
	Latency      float64   `json:"latency"`
	TokenCount   int       `json:"token_count"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	Domain       string    `json:"domain"`
	Status       string    `json:"status" binding:"required"`
	Error        string    `json:"error"`
	Metadata     string    `json:"metadata"`
	RagUsed      bool      `json:"rag_used"`
	Cached       bool      `json:"cached"`
	Intent       string    `json:"intent"`
	Confidence   float64   `json:"confidence"`
}

// InteractionFeedbackInput is used to provide feedback for an interaction
type InteractionFeedbackInput struct {
	InteractionID uuid.UUID `json:"interaction_id" binding:"required"`
	FeedbackScore float64   `json:"feedback_score" binding:"required,min=0,max=1"`
	FeedbackText  string    `json:"feedback_text"`
}

// UploadDocumentInput is used to upload a document to the knowledge base
type UploadDocumentInput struct {
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Domain     string `json:"domain" binding:"required"`
	Source     string `json:"source"`
	SourceType string `json:"source_type"`
	UploadedBy *uint  `json:"uploaded_by"`
	Filename   string `json:"filename"`
	FileType   string `json:"file_type"`
	FileSize   int64  `json:"file_size"`
	Tags       string `json:"tags"`
}

// QueryInput represents a query request
type QueryInput struct {
	Query         string `json:"query" binding:"required"`
	Domain        string `json:"domain"`
	UserID        *uint  `json:"user_id"`
	Limit         int    `json:"limit"`
	UseCache      bool   `json:"use_cache"`
	IncludeSource bool   `json:"include_source"`
}

// TableName sets the table name for the Model struct
func (Model) TableName() string {
	return "ai_models"
}

// TableName sets the table name for the Interaction struct
func (Interaction) TableName() string {
	return "ai_interactions"
}

// TableName sets the table name for the Document struct
func (Document) TableName() string {
	return "rag_documents"
}

// TableName sets the table name for the Chunk struct
func (Chunk) TableName() string {
	return "rag_chunks"
}

// TableName sets the table name for the UserContext struct
func (UserContext) TableName() string {
	return "ai_user_contexts"
}

// BeforeCreate is a GORM hook that runs before creating a new record
func (m *Model) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// BeforeCreate is a GORM hook that runs before creating a new record
func (i *Interaction) BeforeCreate(tx *gorm.DB) (err error) {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	if i.SessionID == "" {
		i.SessionID = uuid.New().String()
	}
	return nil
}

// BeforeCreate is a GORM hook that runs before creating a new record
func (d *Document) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// BeforeCreate is a GORM hook that runs before creating a new record
func (c *Chunk) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (u *UserContext) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
