package cache

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/events"
	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/config"
	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/logger"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var log = logger.NewLogger()

// Custom error types
var (
	ErrCacheNotFound   = errors.New("cache: key not found")
	ErrCacheConnection = errors.New("cache: connection error")
	ErrCacheTimeout    = errors.New("cache: operation timeout")
	ErrInvalidConfig   = errors.New("cache: invalid configuration")
)

// Config holds the configuration for Redis client
type Config struct {
	Addr             string
	Password         string
	DB               int
	PoolSize         int
	MinIdleConns     int
	MaxRetries       int
	ConnTimeout      time.Duration
	OperationTimeout time.Duration
	UseCompression   bool
	DefaultTTL       time.Duration
	MaxKeyLength     int           // Maximum allowed key length
	KeyPrefix        string        // Prefix for all keys
	RetryInterval    time.Duration // Interval between retry attempts
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		PoolSize:         100,
		MinIdleConns:     10,
		MaxRetries:       3,
		ConnTimeout:      5 * time.Second,
		OperationTimeout: 2 * time.Second,
		UseCompression:   false,
		DefaultTTL:       30 * time.Minute,
		MaxKeyLength:     256,
		KeyPrefix:        "compass:",
		RetryInterval:    100 * time.Millisecond,
	}
}

// NewConfigFromEnv creates a Redis config from project configuration
func NewConfigFromEnv(cfg *config.Config) *Config {
	return &Config{
		Addr:             fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:         cfg.Redis.Password,
		DB:               cfg.Redis.DB,
		PoolSize:         100,
		MinIdleConns:     10,
		MaxRetries:       3,
		ConnTimeout:      5 * time.Second,
		OperationTimeout: cfg.Server.Timeout,
		UseCompression:   false,
		DefaultTTL:       30 * time.Minute,
		MaxKeyLength:     256,
		KeyPrefix:        "compass:",
		RetryInterval:    100 * time.Millisecond,
	}
}

// CacheMetrics tracks cache hit/miss statistics with atomic operations
type CacheMetrics struct {
	hits      atomic.Int64
	misses    atomic.Int64
	hitRate   atomic.Int64 // Store as integer (multiply by 100 for percentage)
	lastReset atomic.Int64
	byType    sync.Map // map[string]*TypeMetrics
}

// TypeMetrics tracks metrics for a specific cache type with atomic operations
type TypeMetrics struct {
	hits   atomic.Int64
	misses atomic.Int64
}

// RedisClient wraps the Redis client with additional functionality
type RedisClient struct {
	client    *redis.Client
	metrics   *CacheMetrics
	ttls      sync.Map // map[string]time.Duration
	config    *Config
	closeOnce sync.Once
	health    int32 // 0 = healthy, 1 = unhealthy, using atomic operations
}

// PubSubManager handles pub/sub functionality with listener management
type PubSubManager struct {
	listeners sync.Map
	client    *RedisClient
}

// DashboardEventChannel is the Redis channel for dashboard events
const DashboardEventChannel = "dashboard:events"

// NewRedisClient creates a new Redis client with the provided configuration
func NewRedisClient(cfg *Config) (*RedisClient, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if cfg.Addr == "" {
		return nil, fmt.Errorf("%w: address is required", ErrInvalidConfig)
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnTimeout)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	r := &RedisClient{
		client:  client,
		config:  cfg,
		metrics: &CacheMetrics{},
	}

	// Initialize default TTLs
	r.ttls.Store("default", 30*time.Minute)
	r.ttls.Store("task", time.Hour)
	r.ttls.Store("user", 2*time.Hour)
	r.ttls.Store("project", time.Hour)
	r.ttls.Store("workflow", time.Hour)
	r.ttls.Store("event", 30*time.Minute)
	r.ttls.Store("task_list", 10*time.Minute)
	r.ttls.Store("event_list", 10*time.Minute)

	// Start health check goroutine
	go r.healthCheckLoop()

	return r, nil
}

// healthCheckLoop periodically checks Redis health
func (r *RedisClient) healthCheckLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), r.config.OperationTimeout)
		if err := r.HealthCheck(ctx); err != nil {
			atomic.StoreInt32(&r.health, 1)
			log.Error("Redis health check failed", zap.Error(err))
		} else {
			atomic.StoreInt32(&r.health, 0)
		}
		cancel()
	}
}

// IsHealthy returns whether Redis is currently healthy
func (r *RedisClient) IsHealthy() bool {
	return atomic.LoadInt32(&r.health) == 0
}

// withContext wraps the context with a timeout if none is set
func (r *RedisClient) withContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); !ok {
		return context.WithTimeout(ctx, r.config.OperationTimeout)
	}
	return ctx, func() {}
}

// validateKey checks if the key is valid
func (r *RedisClient) validateKey(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("%w: empty key", ErrInvalidConfig)
	}
	if len(key) > r.config.MaxKeyLength {
		return fmt.Errorf("%w: key too long (max %d characters)", ErrInvalidConfig, r.config.MaxKeyLength)
	}
	return nil
}

// prefixKey adds the configured prefix to the key
func (r *RedisClient) prefixKey(key string) string {
	return r.config.KeyPrefix + key
}

// Get retrieves a value from the cache with proper context handling
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	if err := r.validateKey(key); err != nil {
		return "", err
	}

	if !r.IsHealthy() {
		return "", ErrCacheConnection
	}

	ctx, cancel := r.withContext(ctx)
	defer cancel()

	prefixedKey := r.prefixKey(key)
	val, err := r.client.Get(ctx, prefixedKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("%w: %s", ErrCacheNotFound, key)
		}
		return "", fmt.Errorf("%w: %v", ErrCacheConnection, err)
	}

	if r.config.UseCompression {
		return r.decompress(val)
	}
	return val, nil
}

// Set stores a value in the cache with proper context and compression handling
func (r *RedisClient) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if err := r.validateKey(key); err != nil {
		return err
	}

	if !r.IsHealthy() {
		return ErrCacheConnection
	}

	ctx, cancel := r.withContext(ctx)
	defer cancel()

	if r.config.UseCompression {
		compressed, err := r.compress(value)
		if err != nil {
			return fmt.Errorf("compression failed: %w", err)
		}
		value = compressed
	}

	prefixedKey := r.prefixKey(key)
	return r.client.Set(ctx, prefixedKey, value, ttl).Err()
}

// compress compresses a string using gzip
func (r *RedisClient) compress(data string) (string, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write([]byte(data)); err != nil {
		return "", err
	}

	if err := gz.Close(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// decompress decompresses a gzipped string
func (r *RedisClient) decompress(data string) (string, error) {
	gr, err := gzip.NewReader(strings.NewReader(data))
	if err != nil {
		return "", err
	}
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	if err != nil {
		return "", err
	}

	return string(decompressed), nil
}

// BatchGet retrieves multiple values from the cache in a single operation
func (r *RedisClient) BatchGet(ctx context.Context, keys []string) (map[string]string, error) {
	if !r.IsHealthy() {
		return nil, ErrCacheConnection
	}

	ctx, cancel := r.withContext(ctx)
	defer cancel()

	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		if err := r.validateKey(key); err != nil {
			return nil, err
		}
		prefixedKeys[i] = r.prefixKey(key)
	}

	pipe := r.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd, len(keys))

	for i, prefixedKey := range prefixedKeys {
		cmds[keys[i]] = pipe.Get(ctx, prefixedKey)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("%w: %v", ErrCacheConnection, err)
	}

	result := make(map[string]string)
	for key, cmd := range cmds {
		val, err := cmd.Result()
		if err == nil {
			if r.config.UseCompression {
				val, err = r.decompress(val)
				if err != nil {
					log.Error("Failed to decompress value", zap.String("key", key), zap.Error(err))
					continue
				}
			}
			result[key] = val
		}
	}

	return result, nil
}

// BatchSet stores multiple values in the cache in a single operation
func (r *RedisClient) BatchSet(ctx context.Context, values map[string]string, ttl time.Duration) error {
	if !r.IsHealthy() {
		return ErrCacheConnection
	}

	ctx, cancel := r.withContext(ctx)
	defer cancel()

	pipe := r.client.Pipeline()

	for key, value := range values {
		if err := r.validateKey(key); err != nil {
			return err
		}

		if r.config.UseCompression {
			compressed, err := r.compress(value)
			if err != nil {
				log.Error("Failed to compress value", zap.String("key", key), zap.Error(err))
				continue
			}
			value = compressed
		}

		prefixedKey := r.prefixKey(key)
		pipe.Set(ctx, prefixedKey, value, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCacheConnection, err)
	}

	return nil
}

// Close properly closes the Redis client and stops background tasks
func (r *RedisClient) Close() error {
	var err error
	r.closeOnce.Do(func() {
		err = r.client.Close()
	})
	return err
}

// trackCacheEvent tracks cache hits/misses with atomic operations
func (r *RedisClient) trackCacheEvent(hit bool, cacheType string) {
	if hit {
		r.metrics.hits.Add(1)
	} else {
		r.metrics.misses.Add(1)
	}

	total := r.metrics.hits.Load() + r.metrics.misses.Load()
	if total > 0 {
		hitRate := int64((float64(r.metrics.hits.Load()) / float64(total)) * 100)
		r.metrics.hitRate.Store(hitRate)
	}

	// Update type metrics
	value, _ := r.metrics.byType.LoadOrStore(cacheType, &TypeMetrics{})
	typeMetrics := value.(*TypeMetrics)

	if hit {
		typeMetrics.hits.Add(1)
	} else {
		typeMetrics.misses.Add(1)
	}
}

// GetMetrics returns current cache metrics with additional information
func (r *RedisClient) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})
	typeMetrics := make(map[string]interface{})

	r.metrics.byType.Range(func(key, value interface{}) bool {
		tm := value.(*TypeMetrics)
		typeMetrics[key.(string)] = map[string]interface{}{
			"hits":   tm.hits.Load(),
			"misses": tm.misses.Load(),
		}
		return true
	})

	stats := r.client.PoolStats()
	metrics["hits"] = r.metrics.hits.Load()
	metrics["misses"] = r.metrics.misses.Load()
	metrics["hit_rate"] = float64(r.metrics.hitRate.Load()) / 100.0
	metrics["by_type"] = typeMetrics
	metrics["health"] = r.IsHealthy()
	metrics["pool_stats"] = map[string]interface{}{
		"total_conns": stats.TotalConns,
		"idle_conns":  stats.IdleConns,
		"stale_conns": stats.StaleConns,
	}
	metrics["config"] = map[string]interface{}{
		"compression": r.config.UseCompression,
		"prefix":      r.config.KeyPrefix,
		"max_retries": r.config.MaxRetries,
	}

	return metrics
}

// ResetCacheMetrics resets the cache hit/miss metrics
func (r *RedisClient) ResetCacheMetrics() {
	r.metrics.hits.Store(0)
	r.metrics.misses.Store(0)
	r.metrics.hitRate.Store(0)
	r.metrics.lastReset.Store(time.Now().Unix())
}

// HealthCheck checks if Redis is responding
func (r *RedisClient) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Delete removes values from the cache
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	if !r.IsHealthy() {
		return ErrCacheConnection
	}

	ctx, cancel := r.withContext(ctx)
	defer cancel()

	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		if err := r.validateKey(key); err != nil {
			return err
		}
		prefixedKeys[i] = r.prefixKey(key)
	}

	return r.client.Del(ctx, prefixedKeys...).Err()
}

// ClearByPattern removes all cache entries matching the given pattern
func (r *RedisClient) ClearByPattern(ctx context.Context, pattern string) error {
	if !r.IsHealthy() {
		return ErrCacheConnection
	}

	ctx, cancel := r.withContext(ctx)
	defer cancel()

	prefixedPattern := r.prefixKey(pattern)
	iter := r.client.Scan(ctx, 0, prefixedPattern, 100).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}

	return nil
}

// GenerateCacheKey creates a unique cache key for the given entity
func GenerateCacheKey(entityType string, entityID interface{}, action string) string {
	if entityType == "dashboard" {
		return fmt.Sprintf("dashboard:metrics:%v", entityID)
	}
	if action == "" {
		return fmt.Sprintf("%s:%v", entityType, entityID)
	}
	return fmt.Sprintf("%s:%v:%s", entityType, entityID, action)
}

// CacheResponse is a generic function to cache any serializable response
func (r *RedisClient) CacheResponse(ctx context.Context, key string, ttl time.Duration, cacheType string, fn func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	cachedData, err := r.Get(ctx, key)
	if err != nil {
		log.Error("Error getting from cache", zap.Error(err))
	} else if cachedData != "" {
		// Track cache hit
		r.trackCacheEvent(true, cacheType)
		log.Debug("Cache hit", zap.String("key", key), zap.String("type", cacheType))

		// Deserialize the cached data
		var result interface{}
		if err := json.Unmarshal([]byte(cachedData), &result); err != nil {
			log.Error("Error deserializing cached data", zap.Error(err))
		} else {
			return result, nil
		}
	}

	// Cache miss, execute the function
	r.trackCacheEvent(false, cacheType)
	log.Debug("Cache miss", zap.String("key", key), zap.String("type", cacheType))

	result, err := fn()
	if err != nil {
		return nil, err
	}

	// Don't cache nil results
	if result == nil {
		return nil, nil
	}

	// Serialize and cache the result
	data, err := json.Marshal(result)
	if err != nil {
		log.Error("Error serializing result", zap.Error(err))
		return result, nil
	}

	if err := r.Set(ctx, key, string(data), ttl); err != nil {
		log.Error("Error caching result", zap.Error(err))
	}

	return result, nil
}

// InvalidateCache removes all cache entries for a specific entity
func (r *RedisClient) InvalidateCache(ctx context.Context, entityType string, entityID interface{}) error {
	pattern := fmt.Sprintf("%s:%v*", entityType, entityID)
	return r.ClearByPattern(ctx, pattern)
}

func (r *RedisClient) GetPoolStats() *redis.PoolStats {
	return r.client.PoolStats()
}

func (r *RedisClient) ExportMetrics() map[string]float64 {
	stats := r.GetPoolStats()
	metrics := map[string]float64{
		"cache_hits":       float64(r.metrics.hits.Load()),
		"cache_misses":     float64(r.metrics.misses.Load()),
		"cache_hit_rate":   float64(r.metrics.hitRate.Load()) / 100.0,
		"cache_last_reset": float64(r.metrics.lastReset.Load()),
		"pool_total_conns": float64(stats.TotalConns),
		"pool_idle_conns":  float64(stats.IdleConns),
		"pool_stale_conns": float64(stats.StaleConns),
	}
	return metrics
}

// GetClient returns the underlying Redis client
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// PublishEvent publishes a JSON-encoded event to the specified Redis channel
func (r *RedisClient) PublishEvent(ctx context.Context, channel string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, channel, data).Err()
}

// NewPubSubManager creates a new PubSubManager
func NewPubSubManager(client *RedisClient) *PubSubManager {
	return &PubSubManager{
		client: client,
	}
}

// Subscribe adds a callback function for a specific channel
func (p *PubSubManager) Subscribe(channel string, callback func(interface{}) error) {
	p.listeners.Store(channel, callback)
}

// Unsubscribe removes a callback function for a specific channel
func (p *PubSubManager) Unsubscribe(channel string) {
	p.listeners.Delete(channel)
}

// Notify sends an event to all listeners for a specific channel
func (p *PubSubManager) Notify(channel string, event interface{}) error {
	if callback, ok := p.listeners.Load(channel); ok {
		if cb, ok := callback.(func(interface{}) error); ok {
			return cb(event)
		}
	}
	return nil
}

// PublishEvent publishes an event to a channel and notifies listeners
func (p *PubSubManager) PublishEvent(ctx context.Context, channel string, payload interface{}) error {
	// Publish to Redis
	if err := p.client.PublishEvent(ctx, channel, payload); err != nil {
		return err
	}

	// Notify local listeners
	return p.Notify(channel, payload)
}

// StartListening starts listening for events on a channel
func (p *PubSubManager) StartListening(ctx context.Context, channel string) error {
	pubsub := p.client.GetClient().Subscribe(ctx, channel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			var event interface{}
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				return err
			}
			if err := p.Notify(channel, event); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// PublishDashboardEvent publishes a dashboard event to Redis
func (r *RedisClient) PublishDashboardEvent(ctx context.Context, event *events.DashboardEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, DashboardEventChannel, data).Err()
}

// SubscribeToDashboardEvents subscribes to dashboard events
func (r *RedisClient) SubscribeToDashboardEvents(ctx context.Context, callback func(*events.DashboardEvent) error) error {
	pubsub := r.client.Subscribe(ctx, DashboardEventChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			var event events.DashboardEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				return err
			}
			if err := callback(&event); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// InvalidateDashboardCache invalidates all dashboard cache for a user
func (r *RedisClient) InvalidateDashboardCache(ctx context.Context, userID uuid.UUID) error {
	pattern := fmt.Sprintf("%sdashboard:*:%v", r.config.KeyPrefix, userID)
	log.Info("Invalidating dashboard cache", zap.String("pattern", pattern))

	iter := r.client.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}
	return nil
}
