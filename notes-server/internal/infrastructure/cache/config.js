require('dotenv').config();

const config = {
  host: process.env.REDIS_HOST || 'redis',
  port: parseInt(process.env.REDIS_PORT || '6380'),
  password: process.env.REDIS_PASSWORD || 'test123',
  db: parseInt(process.env.REDIS_DB || '2'),
  keyPrefix: 'compass:notes:',
  maxRetries: 3,
  retryDelay: 100,
  useCompression: false,
  defaultTTL: 30 * 60, 
  healthCheckInterval: 10000, // 10 seconds
  poolSize: 10,
  minIdleConns: 5,
  maxIdleConns: 10,
  connectTimeout: 10000, // 10 seconds
  enableOfflineQueue: true,
  enableReadyCheck: true,
  maxRetriesPerRequest: 3,
  retryStrategy: (times) => {
    if (times > 3) return null;
    return Math.min(times * 100, 3000);
  },
  userTTLs: {
    //notes: 30 * 60,
    //journals: 30 * 60,
    //tags: 30 * 60
  }
};

module.exports = config; 