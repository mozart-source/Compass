package ai

import (
    "context"
    "fmt"
    "time"

    "github.com/ahmedelhadi17776/Compass/Backend_go/pkg/logger"
    "github.com/ahmedelhadi17776/Compass/Backend_go/proto/ai"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

var log = logger.NewLogger()

// Client is a client for the AI service
type Client struct {
    conn   *grpc.ClientConn
    client ai.AIServiceClient
}

// NewClient creates a new AI service client
func NewClient(serverAddr string) (*Client, error) {
    conn, err := grpc.Dial(
        serverAddr,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithBlock(),
        grpc.WithTimeout(5*time.Second),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to connect to AI service: %w", err)
    }

    return &Client{
        conn:   conn,
        client: ai.NewAIServiceClient(conn),
    }, nil
}

// ProcessQuery sends a query to the AI service
func (c *Client) ProcessQuery(ctx context.Context, query string, userID int32, contextData map[string]string, useRAG bool) (*ai.QueryResponse, error) {
    req := &ai.QueryRequest{
        Query:   query,
        UserId:  userID,
        Context: contextData,
        UseRag:  useRAG,
    }

    log.Debug("Sending query to AI service",
        zap.String("query", query),
        zap.Int32("user_id", userID),
        zap.Bool("use_rag", useRAG))

    resp, err := c.client.ProcessQuery(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("AI service query failed: %w", err)
    }

    return resp, nil
}

// GenerateEmbeddings generates embeddings for the given texts
func (c *Client) GenerateEmbeddings(ctx context.Context, texts []string, model string) (*ai.EmbeddingResponse, error) {
    req := &ai.EmbeddingRequest{
        Texts: texts,
        Model: model,
    }

    resp, err := c.client.GenerateEmbeddings(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("embedding generation failed: %w", err)
    }

    return resp, nil
}

// AnalyzeTask analyzes a task and returns insights
func (c *Client) AnalyzeTask(ctx context.Context, taskID int32, title, description, status string, dependencies []int32) (*ai.TaskAnalysisResponse, error) {
    req := &ai.TaskAnalysisRequest{
        TaskId:       taskID,
        Title:        title,
        Description:  description,
        Status:       status,
        Dependencies: dependencies,
    }

    resp, err := c.client.AnalyzeTask(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("task analysis failed: %w", err)
    }

    return resp, nil
}

// GenerateWorkflowRecommendations generates workflow recommendations
func (c *Client) GenerateWorkflowRecommendations(ctx context.Context, userID int32, workflowType string, taskIDs []int32, parameters map[string]string) (*ai.WorkflowResponse, error) {
    req := &ai.WorkflowRequest{
        UserId:       userID,
        WorkflowType: workflowType,
        TaskIds:      taskIDs,
        Parameters:   parameters,
    }

    resp, err := c.client.GenerateWorkflowRecommendations(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("workflow recommendation generation failed: %w", err)
    }

    return resp, nil
}

// AnalyzeProductivity analyzes user productivity
func (c *Client) AnalyzeProductivity(ctx context.Context, userID int32, timePeriod string, completedTasks []*ai.ProductivityRequest_CompletedTask) (*ai.ProductivityResponse, error) {
    req := &ai.ProductivityRequest{
        UserId:         userID,
        TimePeriod:     timePeriod,
        CompletedTasks: completedTasks,
    }

    resp, err := c.client.AnalyzeProductivity(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("productivity analysis failed: %w", err)
    }

    return resp, nil
}

// Close closes the client connection
func (c *Client) Close() error {
    return c.conn.Close()
}