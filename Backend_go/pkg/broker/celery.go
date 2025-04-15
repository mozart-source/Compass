package broker

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/go-redis/redis/v8"
    "github.com/google/uuid"
    "github.com/ahmedelhadi17776/Compass/Backend_go/pkg/logger"
    "go.uber.org/zap"
)

var log = logger.NewLogger()
// TaskStatus represents the status of a task
type TaskStatus string

const (
    TaskStatusPending   TaskStatus = "PENDING"
    TaskStatusStarted   TaskStatus = "STARTED"
    TaskStatusSuccess   TaskStatus = "SUCCESS"
    TaskStatusFailure   TaskStatus = "FAILURE"
    TaskStatusRevoked   TaskStatus = "REVOKED"
    TaskStatusRetry     TaskStatus = "RETRY"
)

// TaskResult represents the result of a task
type TaskResult struct {
    ID        string          `json:"id"`
    Status    TaskStatus      `json:"status"`
    Result    json.RawMessage `json:"result,omitempty"`
    Error     string          `json:"error,omitempty"`
    Traceback string          `json:"traceback,omitempty"`
    CreatedAt time.Time       `json:"created_at"`
    StartedAt *time.Time      `json:"started_at,omitempty"`
    EndedAt   *time.Time      `json:"ended_at,omitempty"`
}

// TaskBroker handles task queue operations
type TaskBroker struct {
    redis      *redis.Client
    queueName  string
    resultsTTL time.Duration
}

// NewTaskBroker creates a new task broker
func NewTaskBroker(redisURL, queueName string, resultsTTL time.Duration) (*TaskBroker, error) {
    opts, err := redis.ParseURL(redisURL)
    if err != nil {
        return nil, fmt.Errorf("invalid Redis URL: %w", err)
    }

    client := redis.NewClient(opts)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if _, err := client.Ping(ctx).Result(); err != nil {
        return nil, fmt.Errorf("failed to connect to Redis: %w", err)
    }

    return &TaskBroker{
        redis:      client,
        queueName:  queueName,
        resultsTTL: resultsTTL,
    }, nil
}

// EnqueueTask adds a task to the queue
func (b *TaskBroker) EnqueueTask(ctx context.Context, taskName string, args interface{}, kwargs map[string]interface{}) (string, error) {
    taskID := uuid.New().String()

    // Create task message in Celery-compatible format
    taskMessage := map[string]interface{}{
        "id":      taskID,
        "task":    taskName,
        "args":    args,
        "kwargs":  kwargs,
        "retries": 0,
        "eta":     nil,
    }

    // Serialize task message
    messageBytes, err := json.Marshal(taskMessage)
    if err != nil {
        return "", fmt.Errorf("failed to serialize task message: %w", err)
    }

    // Store initial task status
    initialResult := TaskResult{
        ID:        taskID,
        Status:    TaskStatusPending,
        CreatedAt: time.Now(),
    }

    initialResultBytes, err := json.Marshal(initialResult)
    if err != nil {
        return "", fmt.Errorf("failed to serialize initial result: %w", err)
    }

    // Store task result with TTL
    resultKey := fmt.Sprintf("celery-task-meta-%s", taskID)
    if err := b.redis.Set(ctx, resultKey, initialResultBytes, b.resultsTTL).Err(); err != nil {
        return "", fmt.Errorf("failed to store initial task result: %w", err)
    }

    // Push task to queue
    if err := b.redis.LPush(ctx, b.queueName, messageBytes).Err(); err != nil {
        return "", fmt.Errorf("failed to enqueue task: %w", err)
    }

    log.Info("Task enqueued", 
        zap.String("task_id", taskID), 
        zap.String("task_name", taskName),
        zap.String("queue", b.queueName))

    return taskID, nil
}

// GetTaskResult retrieves the result of a task
func (b *TaskBroker) GetTaskResult(ctx context.Context, taskID string) (*TaskResult, error) {
    resultKey := fmt.Sprintf("celery-task-meta-%s", taskID)
    resultBytes, err := b.redis.Get(ctx, resultKey).Bytes()
    if err != nil {
        if err == redis.Nil {
            return nil, fmt.Errorf("task result not found")
        }
        return nil, fmt.Errorf("failed to get task result: %w", err)
    }

    var result TaskResult
    if err := json.Unmarshal(resultBytes, &result); err != nil {
        return nil, fmt.Errorf("failed to deserialize task result: %w", err)
    }

    return &result, nil
}

// RevokeTask revokes a task
func (b *TaskBroker) RevokeTask(ctx context.Context, taskID string) error {
    // Get current task result
    result, err := b.GetTaskResult(ctx, taskID)
    if err != nil {
        return err
    }

    // Update task status to revoked
    result.Status = TaskStatusRevoked
    now := time.Now()
    result.EndedAt = &now

    // Serialize updated result
    resultBytes, err := json.Marshal(result)
    if err != nil {
        return fmt.Errorf("failed to serialize updated result: %w", err)
    }

    // Store updated task result
    resultKey := fmt.Sprintf("celery-task-meta-%s", taskID)
    if err := b.redis.Set(ctx, resultKey, resultBytes, b.resultsTTL).Err(); err != nil {
        return fmt.Errorf("failed to store updated task result: %w", err)
    }

    // Add task ID to revoked set
    revokedKey := "celery-revoked"
    if err := b.redis.SAdd(ctx, revokedKey, taskID).Err(); err != nil {
        return fmt.Errorf("failed to add task to revoked set: %w", err)
    }

    log.Info("Task revoked", zap.String("task_id", taskID))

    return nil
}

// Close closes the broker connection
func (b *TaskBroker) Close() error {
    return b.redis.Close()
}