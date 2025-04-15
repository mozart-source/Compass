const { RedisPubSub } = require('graphql-redis-subscriptions');
const redisConfig = require('./config');

const pubsub = new RedisPubSub({
  connection: {
    host: redisConfig.host || 'localhost',
    port: redisConfig.port || 6380,
    password: redisConfig.password || '',
    db: redisConfig.db || 2
  },
  retryStrategy: times => Math.min(times * 50, 2000)
});

module.exports = { pubsub }; 