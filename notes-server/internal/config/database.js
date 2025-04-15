const { connectDB } = require('../infrastructure/persistence/mongodb/connection');
const RedisService = require('../infrastructure/cache/redisService');
const redisConfig = require('../infrastructure/cache/config');
const { logger } = require('../../pkg/utils/logger');

const initializeDatabases = async () => {
  try {
    // Connect to MongoDB
    await connectDB();
    logger.info('Connected to MongoDB');
    
    // Initialize Redis
    const redisClient = new RedisService(redisConfig);
    logger.info('Redis client initialized');
    
    // Make Redis client available globally
    global.redisClient = redisClient;
    
    // Wait for Redis to be ready
    await redisClient.client.ping();
    logger.info('Redis connection verified');

    return redisClient;
  } catch (error) {
    logger.error('Database initialization failed', {
      error: error.stack,
      message: error.message
    });
    throw error;
  }
};

module.exports = initializeDatabases; 