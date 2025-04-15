const { logger } = require('../../pkg/utils/logger');
const dashboardSubscriber = require('../infrastructure/cache/dashboardSubscriber');

const setupShutdownHandlers = (server, redisClient) => {
  const gracefulShutdown = async () => {
    logger.info('Starting graceful shutdown...');
    let shutdownTimeout = setTimeout(() => {
      logger.error('Graceful shutdown timed out, forcing exit.');
      process.exit(1);
    }, 5000); // 5 seconds fallback
    try {
      // Unsubscribe from dashboard events
      await dashboardSubscriber.unsubscribe();
      logger.info('Dashboard subscriber unsubscribed');

      // Close Redis connection
      if (redisClient) {
        await redisClient.close();
        logger.info('Redis connection closed');
      }
      // Close server
      if (server) {
        server.close(() => {
          logger.info('Express server closed');
          clearTimeout(shutdownTimeout);
        });
      } else {
        clearTimeout(shutdownTimeout);
      }
    } catch (err) {
      logger.error('Error during graceful shutdown', { error: err.message });
      clearTimeout(shutdownTimeout);
      process.exit(1);
    }
  };

  // Handle various shutdown signals
  process.on('SIGTERM', gracefulShutdown);
  process.on('SIGINT', gracefulShutdown);

  // Handle uncaught errors
  process.on('uncaughtException', (error) => {
    logger.error('Uncaught Exception', {
      error: error.stack,
      message: error.message
    });
    gracefulShutdown();
  });

  process.on('unhandledRejection', (error) => {
    logger.error('Unhandled Rejection', {
      error: error.stack,
      message: error.message
    });
    gracefulShutdown();
  });
};

module.exports = setupShutdownHandlers;