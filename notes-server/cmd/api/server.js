require('dotenv').config();

const express = require('express');
const { logger } = require('../../pkg/utils/logger');
const { createRateLimiter } = require('../../internal/api/middleware');
const healthRoutes = require('../../internal/api/routes/health');
const graphqlRoutes = require('../../internal/api/routes/graphql');
const dashboardRoutes = require('../../internal/api/routes/dashboard');
const { configureServer } = require('../../internal/config/server');
const setupShutdownHandlers = require('../../internal/config/shutdown');
const initializeDatabases = require('../../internal/config/database');
const dashboardSubscriber = require('../../internal/infrastructure/cache/dashboardSubscriber');
const { useServer } = require('graphql-ws/lib/use/ws');
const { createServer } = require('http');
const { WebSocketServer } = require('ws');
const schema = require('../../internal/api/graphql');

const app = express();
let server;
let redisClient;
let wsServer;
let wsCleanup;

// Configure server middleware and error handling
configureServer(app);

// Initialize server
const initializeServer = async () => {
  try {
    logger.info('Initializing server...');

    // Initialize databases
    redisClient = await initializeDatabases();

    // Initialize rate limiter with Redis client
    const { limiter, rateLimitHeaders, getRateLimitInfo } = createRateLimiter(redisClient);
    logger.info('Rate limiter initialized');

    // Apply rate limiting middleware globally
    app.use(limiter);
    app.use(rateLimitHeaders);

    // Routes
    app.use('/health', healthRoutes);
    app.use('/graphql', graphqlRoutes);
    app.use('/api/dashboard', dashboardRoutes);

    // Initialize dashboard subscriber
    await dashboardSubscriber.subscribe();
    logger.info('Dashboard subscriber initialized');

    // Add rate limit info endpoint
    app.get('/rate-limit-info', async (req, res) => {
      try {
        const info = await getRateLimitInfo(req);
        logger.debug('Rate limit info requested', { info });
        res.json(info);
      } catch (error) {
        logger.error('Error getting rate limit info', { error: error.message });
        res.status(500).json({
          success: false,
          message: 'Error getting rate limit info',
          error: error.message
        });
      }
    });

    const PORT = process.env.PORT || 5000;

    // Create HTTP server
    const httpServer = createServer(app);

    // Set up WebSocket server for GraphQL subscriptions
    wsServer = new WebSocketServer({
      server: httpServer,
      path: '/graphql',
    });

    wsCleanup = useServer({
      schema,
      context: async (ctx, msg, args) => {
        // Extract token from connection params or query parameters
        const token = ctx.connectionParams?.Authorization || 
                     ctx.extra?.request?.headers?.authorization ||
                     new URLSearchParams(ctx.extra?.request?.url?.split('?')[1] || '').get('token');
        
        return { 
          token,
          user: ctx.extra.request.user // Will be set by auth middleware if available
        };
      },
      onConnect: async (ctx) => {
        logger.info('WebSocket GraphQL client connected');
        return true;
      },
      onDisconnect: async (ctx, code, reason) => {
        logger.info('WebSocket GraphQL client disconnected', { code, reason });
      },
      onError: (ctx, msg, errors) => {
        logger.error('GraphQL WebSocket error', { errors });
      }
    }, wsServer);

    // Start server
    server = httpServer.listen(PORT, () => {
      logger.info(`Server running on port ${PORT}`);
      logger.info(`GraphQL endpoint: http://localhost:${PORT}/graphql`);
      logger.info(`GraphQL WebSocket: ws://localhost:${PORT}/graphql`);
    });

    // Setup shutdown handlers
    setupShutdownHandlers(server, redisClient);
    
    // Add WebSocket server cleanup on shutdown
    const cleanup = async (signal) => {
      logger.info(`Received ${signal}, starting graceful shutdown...`);
      if (wsCleanup) {
        await wsCleanup.dispose();
        logger.info('WebSocket server cleaned up');
      }
    };
    
    process.on('SIGTERM', cleanup);
    process.on('SIGINT', cleanup);

  } catch (error) {
    logger.error('Failed to initialize server', {
      error: error.stack,
      message: error.message
    });
    process.exit(1);
  }
};

initializeServer();
