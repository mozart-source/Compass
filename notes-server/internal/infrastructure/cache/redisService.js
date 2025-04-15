const Redis = require('ioredis');
const { logger } = require('../../../pkg/utils/logger');
const { DatabaseError } = require('../../../pkg/utils/errorHandler');
const { compress, decompress, safeStringify, safeParse, toBase64, fromBase64 } = require('./utils/compressionUtils');
const { generateEntityKey, generateListKey, generateTagSetKey, paramStringify } = require('./utils/keyUtils');

class RedisService {
  constructor(config) {
    this.config = {
      host: config.host || 'localhost',
      port: config.port || 6380,
      password: config.password || '',
      db: config.db || 2,
      keyPrefix: 'compass:notes:',
      maxRetries: 3,
      retryDelay: 100,
      useCompression: false,
      defaultTTL: 30 * 60,
      ...config
    };

    logger.info('Initializing Redis client', {
      host: this.config.host,
      port: this.config.port,
      db: this.config.db
    });

    this.client = new Redis({
      host: this.config.host,
      port: this.config.port,
      password: this.config.password,
      db: this.config.db,
      retryStrategy: (times) => {
        const delay = Math.min(times * 50, 2000);
        logger.info('Redis reconnecting...', { attempt: times, delay });
        return delay;
      },
      maxRetriesPerRequest: 3
    });

    this.client.on('error', (error) => {
      logger.error('Redis client error:', { error: error.message });
    });
    this.client.on('connect', () => {
      logger.info('Redis client connected');
    });
    this.client.on('ready', () => {
      logger.info('Redis client ready');
    });
    this.client.on('reconnecting', () => {
      logger.info('Redis client reconnecting');
    });
    this.client.on('end', () => {
      logger.info('Redis client connection ended');
    });

    this.metrics = {
      hits: 0,
      misses: 0,
      byType: new Map()
    };
    this.health = true;
    this.startHealthCheck();
  }

  async startHealthCheck() {
    setInterval(async () => {
      try {
        await this.client.ping();
        this.health = true;
      } catch (error) {
        this.health = false;
        logger.error('Redis health check failed:', { error: error.message });
      }
    }, 10000);
  }

  async isHealthy() {
    try {
      if (!this.client.status === 'ready') {
        return false;
      }
      await this.client.ping();
      return true;
    } catch (error) {
      logger.error('Redis health check failed:', { error: error.message });
      return false;
    }
  }

  getUserTTL(userId) {
    if (this.config.userTTLs && this.config.userTTLs[userId]) {
      return this.config.userTTLs[userId];
    }
    return this.config.defaultTTL;
  }

  async getEntity(entityType, id) {
    const key = generateEntityKey(this.config.keyPrefix, entityType, id);
    return this.get(key);
  }

  async setEntity(entityType, id, value, tags, userId, ttl = null) {
    const key = generateEntityKey(this.config.keyPrefix, entityType, id);
    const effectiveTTL = ttl || this.getUserTTL(userId);
    return this.cacheWithTags(key, value, tags, effectiveTTL);
  }

  async getList(key) {
    return this.get(key);
  }

  async setList(key, value, tags, userId, ttl = null) {
    const effectiveTTL = ttl || this.getUserTTL(userId);
    return this.cacheWithTags(key, value, tags, effectiveTTL);
  }

  async invalidateByPattern(pattern) {
    return this.clearByPattern(pattern);
  }

  async get(key) {
    try {
      const value = await this.client.get(key);
      return value ? safeParse(value) : null;
    } catch (error) {
      logger.error('Redis get error:', { key, error: error.message });
      throw new DatabaseError(`Failed to get value from Redis: ${error.message}`);
    }
  }

  async set(key, value, ttl = null) {
    try {
      const stringValue = safeStringify(value);
      if (ttl) {
        await this.client.setex(key, ttl, stringValue);
      } else {
        await this.client.set(key, stringValue);
      }
    } catch (error) {
      logger.error('Redis set error:', { key, error: error.message });
      throw new DatabaseError(`Failed to set value in Redis: ${error.message}`);
    }
  }

  async removeKeyFromAllTagSets(key) {
    try {
      let cursor = '0';
      do {
        const [nextCursor, tagSets] = await this.client.scan(cursor, 'MATCH', generateTagSetKey(this.config.keyPrefix, '*'), 'COUNT', 100);
        cursor = nextCursor;
        for (const tagSet of tagSets) {
          await this.client.srem(tagSet, key);
        }
      } while (cursor !== '0');
    } catch (error) {
      logger.error('Failed to remove key from tag sets', { key, error: error.message });
    }
  }

  async del(key) {
    try {
      await this.client.del(key);
      await this.removeKeyFromAllTagSets(key);
    } catch (error) {
      logger.error('Redis delete error:', { key, error: error.message });
      throw new DatabaseError(`Failed to delete value from Redis: ${error.message}`);
    }
  }

  async clearByPattern(pattern) {
    try {
      let cursor = '0';
      let keys = [];
      do {
        const [nextCursor, foundKeys] = await this.client.scan(cursor, 'MATCH', pattern, 'COUNT', 100);
        cursor = nextCursor;
        keys = keys.concat(foundKeys);
      } while (cursor !== '0');
      if (keys.length > 0) {
        if (keys.length > 1000) {
          logger.warn('Large number of keys to delete in clearByPattern', { pattern, count: keys.length });
        }
        await Promise.all(keys.map(async (key) => {
          await this.client.del(key);
          await this.removeKeyFromAllTagSets(key);
        }));
      }
    } catch (error) {
      logger.error('Redis clear pattern error:', { pattern, error: error.message });
      throw new DatabaseError(`Failed to clear pattern from Redis: ${error.message}`);
    }
  }

  trackCacheEvent(hit, type) {
    if (hit) {
      this.metrics.hits++;
    } else {
      this.metrics.misses++;
    }
    const typeMetrics = this.metrics.byType.get(type) || { hits: 0, misses: 0 };
    if (hit) {
      typeMetrics.hits++;
    } else {
      typeMetrics.misses++;
    }
    this.metrics.byType.set(type, typeMetrics);
  }

  getMetrics() {
    const total = this.metrics.hits + this.metrics.misses;
    const hitRate = total > 0 ? (this.metrics.hits / total) * 100 : 0;
    const metrics = {
      hits: this.metrics.hits,
      misses: this.metrics.misses,
      hitRate: hitRate.toFixed(2) + '%',
      byType: Object.fromEntries(this.metrics.byType),
      health: this.isHealthy()
    };
    logger.debug('Cache metrics', { metrics });
    return metrics;
  }

  async cacheWithTags(key, value, tags, ttl = this.config.defaultTTL) {
    try {
      await this.set(key, value, ttl);
      const tagSets = tags.map(tag => generateTagSetKey(this.config.keyPrefix, tag));
      await Promise.all(tagSets.map(tagSet => this.client.sadd(tagSet, key)));
      return true;
    } catch (error) {
      logger.error('Cache with tags error:', { error: error.message });
      return false;
    }
  }

  async invalidateByTags(tags) {
    try {
      const tagSets = tags.map(tag => generateTagSetKey(this.config.keyPrefix, tag));
      const keys = await Promise.all(tagSets.map(tagSet => this.client.smembers(tagSet)));
      const uniqueKeys = [...new Set(keys.flat())];
      await Promise.all([
        ...uniqueKeys.map(key => this.del(key)),
        ...tagSets.map(tagSet => this.client.del(tagSet))
      ]);
      return true;
    } catch (error) {
      logger.error('Invalidate by tags error:', { error: error.message });
      return false;
    }
  }

  async close() {
    try {
      await this.client.quit();
      logger.info('Redis connection closed');
    } catch (error) {
      logger.error('Error closing Redis connection:', { error: error.message });
      throw new DatabaseError(`Failed to close Redis connection: ${error.message}`);
    }
  }

  generateKey(entityType, entityId, action = '') {
    return generateEntityKey(this.config.keyPrefix, entityType, entityId, action);
  }

  generateListKey(userId, entityType, params = {}) {
    return generateListKey(userId, entityType, params);
  }
}

module.exports = RedisService; 