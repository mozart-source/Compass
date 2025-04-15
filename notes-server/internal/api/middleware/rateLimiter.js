const rateLimit = require('express-rate-limit');
const { RedisStore } = require('rate-limit-redis');
const { DatabaseError } = require('../../../pkg/utils/errorHandler');
const { logger } = require('../../../pkg/utils/logger');

// Reuse the existing Redis client from global.redisClient
const createRateLimiter = (redisClient) => {
  if (!redisClient) {
    throw new Error('Redis client is required for rate limiting');
  }

  // Key generator function
  const keyGenerator = (req) => {
    // Use IP and user ID if available for better rate limiting
    const key = req.user ? `${req.ip}-${req.user.id}` : req.ip;
    // Add organization ID if available
    return req.headers['x-organization-id'] ? `${key}-${req.headers['x-organization-id']}` : key;
  };

  const windowMs = parseInt(process.env.RATE_LIMIT_WINDOW_MS, 10) || 15 * 60 * 1000;
  const max = parseInt(process.env.RATE_LIMIT_MAX_REQUESTS, 10) || 100;
  const blockDuration = parseInt(process.env.RATE_LIMIT_BLOCK_DURATION, 10) || 60 * 60;

  const limiter = rateLimit({
    store: new RedisStore({
      sendCommand: (...args) => redisClient.client.call(...args),
      prefix: 'rate-limit:',
      windowMs,
      max,
      blockDuration,
      onError: (err) => {
        logger.error('Rate limit store error:', err);
        throw new DatabaseError(`Rate limit store error: ${err.message}`);
      }
    }),
    windowMs,
    max,
    standardHeaders: true,
    legacyHeaders: false,
    message: {
      success: false,
      message: 'Too many requests, please try again later.',
      errors: [{
        message: 'Rate limit exceeded',
        code: 'RATE_LIMIT_EXCEEDED',
        retryAfter: 15 * 60 
      }]
    },
    skip: (req) => {
      // Skip rate limiting for health checks and OPTIONS requests
      return req.path === '/health' || req.method === 'OPTIONS';
    },
    keyGenerator,
    handler: (req, res) => {
      res.status(429).json({
        success: false,
        message: 'Too many requests, please try again later.',
        errors: [{
          message: 'Rate limit exceeded',
          code: 'RATE_LIMIT_EXCEEDED',
          retryAfter: 15 * 60
        }]
      });
    }
  });

  // Helper function to get rate limit info
  const getRateLimitInfo = async (req) => {
    try {
      const key = keyGenerator(req);
      const windowMs = limiter.windowMs;
      const max = limiter.max;
      
      // Get current count from Redis
      const count = await redisClient.client.get(`rate-limit:${key}`);
      const ttl = await redisClient.client.ttl(`rate-limit:${key}`);
      
      return {
        key,
        windowMs,
        max,
        current: parseInt(count) || 0,
        remaining: max - (parseInt(count) || 0),
        reset: ttl > 0 ? Math.ceil(ttl / 1000) : 0
      };
    } catch (error) {
      logger.error('Error getting rate limit info:', error);
      return null;
    }
  };

  // Add rate limit info to response headers
  const rateLimitHeaders = async (req, res, next) => {
    try {
      const info = await getRateLimitInfo(req);
      if (info) {
        res.set('X-RateLimit-Limit', info.max);
        res.set('X-RateLimit-Remaining', info.remaining);
        res.set('X-RateLimit-Reset', info.reset);
      }
      next();
    } catch (error) {
      logger.error('Error setting rate limit headers:', error);
      next();
    }
  };

  return {
    limiter,
    getRateLimitInfo,
    rateLimitHeaders
  };
};

module.exports = createRateLimiter;