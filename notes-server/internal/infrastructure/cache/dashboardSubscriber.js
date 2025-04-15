const Redis = require('ioredis');
const { logger } = require('../../../pkg/utils/logger');
const redisConfig = require('./config');
const RedisService = require('./redisService');

// Dashboard event types
const EVENT_TYPES = {
  METRICS_UPDATE: 'metrics_update',
  CACHE_INVALIDATE: 'cache_invalidate'
};

// Dashboard event channel - must match the Go backend
const DASHBOARD_EVENT_CHANNEL = 'dashboard:events';

const redisClient = new RedisService(redisConfig);

class DashboardEventSubscriber {
  constructor() {
    this.redisClient = new RedisService(redisConfig);
    this.subscriber = null;
    this.isSubscribed = false;
  }

  async subscribe() {
    if (this.isSubscribed) {
      return;
    }

    this.subscriber = this.redisClient.client.duplicate();
    await this.subscriber.subscribe('dashboard:events');

    this.subscriber.on('message', async (channel, message) => {
      try {
        const event = JSON.parse(message);
        
        // Handle events from Python and Go backends
        if (event.event_type === 'metrics_update') {
          await this.handleMetricsUpdate(event.user_id, event.details);
        } else if (event.event_type === 'cache_invalidate') {
          await this.handleCacheInvalidate(event.user_id);
        }
      } catch (error) {
        logger.error('Error handling dashboard event', { error: error.message });
      }
    });

    this.isSubscribed = true;
    logger.info('Subscribed to dashboard events');
  }

  async handleMetricsUpdate(userId, details) {
    try {
      // Invalidate local cache when metrics are updated
      await this.redisClient.del(`compass:notes:dashboard:metrics:${userId}`);
      logger.info('Invalidated dashboard metrics cache', { userId });
    } catch (error) {
      logger.error('Error handling metrics update', { error: error.message });
    }
  }

  async handleCacheInvalidate(userId) {
    try {
      // Invalidate all user-related caches
      await this.redisClient.clearByPattern(`compass:notes:dashboard:metrics:${userId}`);
      logger.info('Invalidated dashboard caches', { userId });
    } catch (error) {
      logger.error('Error handling cache invalidation', { error: error.message });
    }
  }

  async unsubscribe() {
    if (this.subscriber) {
      await this.subscriber.unsubscribe('dashboard:events');
      await this.subscriber.quit();
      this.subscriber = null;
      this.isSubscribed = false;
      logger.info('Unsubscribed from dashboard events');
    }
  }
}

module.exports = new DashboardEventSubscriber();