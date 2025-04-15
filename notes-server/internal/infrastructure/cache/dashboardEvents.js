const { pubsub } = require('./pubsub');
const RedisService = require('./redisService');
const redisConfig = require('./config');
const { logger } = require('../../../pkg/utils/logger');

const redisClient = new RedisService(redisConfig);

const DASHBOARD_EVENT_CHANNEL = 'dashboard:events';
const DASHBOARD_EVENT_TYPES = {
  METRICS_UPDATE: 'metrics_update',
  CACHE_INVALIDATE: 'cache_invalidate'
};

class DashboardEventPublisher {
  constructor() {
    this.redisClient = new RedisService(redisConfig);
  }

  /**
   * Publish a metrics update event
   * @param {string} userId - The user ID
   * @param {string} entityId - Optional entity ID (defaults to userId)
   * @param {Object} metrics - Optional metrics data
   * @param {Object} details - Additional details
   * @returns {Promise<Object>} - The published event
   */
  async publishMetricsUpdate(userId, entityId = null, metrics = null, details = {}) {
    if (!userId) {
      logger.warn('Cannot publish metrics update without userId');
      throw new Error('userId is required');
    }

    const event = {
      event_type: 'metrics_update',
      user_id: userId,
      entity_id: entityId || userId,
      timestamp: new Date().toISOString(),
      details: {
        ...details,
        metrics: metrics || {}
      }
    };

    try {
      // Publish to Redis for Python and Go backends to consume
      await this.redisClient.client.publish(DASHBOARD_EVENT_CHANNEL, JSON.stringify(event));

      // Also invalidate local cache
      await this.redisClient.del(`compass:notes:dashboard:metrics:${userId}`);

      logger.info('Published dashboard metrics update event', { userId, entityId });
      return event;
    } catch (error) {
      logger.error('Failed to publish dashboard event', { error: error.message, userId });
      throw error;
    }
  }

  /**
   * Publish a cache invalidation event
   * @param {string} userId - The user ID
   * @param {string} entityId - Optional entity ID (defaults to userId)
   * @param {Object} details - Additional details
   * @returns {Promise<Object>} - The published event
   */
  async publishCacheInvalidation(userId, entityId = null, details = {}) {
    if (!userId) {
      logger.warn('Cannot publish cache invalidation without userId');
      throw new Error('userId is required');
    }

    const event = {
      event_type: 'cache_invalidate',
      user_id: userId,
      entity_id: entityId || userId,
      timestamp: new Date().toISOString(),
      details
    };

    try {
      // Publish to Redis for Python and Go backends to consume
      await this.redisClient.client.publish(DASHBOARD_EVENT_CHANNEL, JSON.stringify(event));

      logger.info('Published dashboard cache invalidation event', { userId, entityId });
      return event;
    } catch (error) {
      logger.error('Failed to publish cache invalidation event', { error: error.message, userId });
      throw error;
    }
  }
}

// Create a singleton instance
const dashboardEvents = new DashboardEventPublisher();

module.exports = {
  dashboardEvents,
  DASHBOARD_EVENT_TYPES
};