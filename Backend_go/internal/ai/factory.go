package ai

import (
	"fmt"
	"sync"

	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/config"

	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// Factory creates and manages AI service clients
type Factory struct {
	config  *config.Config
	clients map[string]*Client
	mutex   sync.RWMutex
}

// NewFactory creates a new AI service factory
func NewFactory(config *config.Config) *Factory {
	return &Factory{
		config:  config,
		clients: make(map[string]*Client),
	}
}

// GetClient retrieves a client by name
func (f *Factory) GetClient(name string) (*Client, bool) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	client, exists := f.clients[name]
	return client, exists
}

// SetClient stores a client with the given name
func (f *Factory) SetClient(name string, client *Client) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.clients[name] = client
}

// GetOrCreateClient gets an existing client or creates a new one
func (f *Factory) GetOrCreateClient(name, serverAddr string) (*Client, error) {
	if client, exists := f.GetClient(name); exists {
		return client, nil
	}

	client, err := NewClient(serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create client %s: %w", name, err)
	}

	f.SetClient(name, client)
	return client, nil
}

// CloseAll closes all managed clients
func (f *Factory) CloseAll() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	var lastErr error
	for name, client := range f.clients {
		if err := client.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close client %s: %w", name, err)
		}
		delete(f.clients, name)
	}
	return lastErr
}

// CacheMetrics tracks cache hit/miss statistics
type CacheMetrics struct {
	Hits      int64
	Misses    int64
	HitRate   float64
	LastReset int64
	ByType    map[string]TypeMetrics
}

// TypeMetrics tracks metrics for a specific cache type
type TypeMetrics struct {
	Hits   int64
	Misses int64
}

// RedisClient wraps the Redis client with additional functionality
type RedisClient struct {
	client  *redis.Client
	metrics CacheMetrics
	ttls    map[string]int
}

// NewRedisClient creates a new Redis client
func NewRedisClient(addr, password string, db int) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     100,
		MinIdleConns: 10,
		MaxRetries:   3,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{
		client: client,
		metrics: CacheMetrics{
			LastReset: time.Now().Unix(),
			ByType:    make(map[string]TypeMetrics),
		},
		ttls: map[string]int{
			"default":    1800, // 30 minutes
			"task":       3600, // 1 hour
			"user":       7200, // 2 hours
			"project":    3600, // 1 hour
			"workflow":   3600, // 1 hour
			"event":      1800, // 30 minutes
			"task_list":  600,  // 10 minutes
			"event_list": 600,  // 10 minutes
		},
	}, nil
}

// Get retrieves a value from the cache
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", err
	}
	return val, nil
}

// Set stores a value in the cache with the specified TTL
func (r *RedisClient) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes a value from the cache
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// GenerateCacheKey creates a unique cache key for the given entity
func GenerateCacheKey(entityType string, entityID interface{}, action string) string {
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

// ClearByPattern removes all cache entries matching the given pattern
func (r *RedisClient) ClearByPattern(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return r.Delete(ctx, keys...)
	}

	return nil
}

// GetCacheStats returns cache statistics
func (r *RedisClient) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	_, err := r.client.Info(ctx, "stats").Result()
	if err != nil {
		return nil, err
	}

	// Parse Redis info
	redisHitRate := 0.0
	// In a real implementation, parse the info string to extract hit rate

	// Calculate application hit rate
	appHitRate := float64(r.metrics.Hits) / float64(r.metrics.Hits+r.metrics.Misses)
	if r.metrics.Hits+r.metrics.Misses == 0 {
		appHitRate = 0
	}

	// Prepare type metrics
	typeMetrics := make(map[string]map[string]interface{})
	for cacheType, metrics := range r.metrics.ByType {
		typeHitRate := float64(metrics.Hits) / float64(metrics.Hits+metrics.Misses)
		if metrics.Hits+metrics.Misses == 0 {
			typeHitRate = 0
		}

		typeMetrics[cacheType] = map[string]interface{}{
			"hits":     metrics.Hits,
			"misses":   metrics.Misses,
			"hit_rate": typeHitRate,
		}
	}

	return map[string]interface{}{
		"redis": map[string]interface{}{
			"hit_rate":   redisHitRate,
			"total_keys": r.client.DBSize(ctx).Val(),
		},
		"app": map[string]interface{}{
			"hits":           r.metrics.Hits,
			"misses":         r.metrics.Misses,
			"hit_rate":       appHitRate,
			"tracking_since": time.Unix(r.metrics.LastReset, 0).Format(time.RFC3339),
			"by_type":        typeMetrics,
		},
	}, nil
}

// ResetCacheMetrics resets the cache hit/miss metrics
func (r *RedisClient) ResetCacheMetrics() {
	r.metrics = CacheMetrics{
		Hits:      0,
		Misses:    0,
		HitRate:   0,
		LastReset: time.Now().Unix(),
		ByType:    make(map[string]TypeMetrics),
	}
}

// trackCacheEvent tracks a cache hit or miss event
func (r *RedisClient) trackCacheEvent(hit bool, cacheType string) {
	if hit {
		r.metrics.Hits++
	} else {
		r.metrics.Misses++
	}

	// Update hit rate
	total := r.metrics.Hits + r.metrics.Misses
	if total > 0 {
		r.metrics.HitRate = float64(r.metrics.Hits) / float64(total)
	}

	// Initialize cache type if not exists
	if _, ok := r.metrics.ByType[cacheType]; !ok {
		r.metrics.ByType[cacheType] = TypeMetrics{}
	}

	// Update type metrics
	typeMetrics := r.metrics.ByType[cacheType]
	if hit {
		typeMetrics.Hits++
	} else {
		typeMetrics.Misses++
	}
	r.metrics.ByType[cacheType] = typeMetrics
}
