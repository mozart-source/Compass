const express = require('express');
const cors = require('cors');
const { logger, requestLogger } = require('../../pkg/utils/logger');
require('dotenv').config();

const config = {
  port: process.env.PORT || 5000,
  nodeEnv: process.env.NODE_ENV || 'development',
  mongodb: {
    uri: process.env.MONGODB_URI || 'mongodb+srv://ahmedelhadi1777:fb5OpNipjvS65euk@cluster0.ojy4aft.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0',
    options: {
      maxPoolSize: parseInt(process.env.MONGODB_MAX_POOL_SIZE) || 50,
      minPoolSize: parseInt(process.env.MONGODB_MIN_POOL_SIZE) || 10,
      maxIdleTimeMS: parseInt(process.env.MONGODB_MAX_IDLE_TIME_MS) || 30000,
      connectTimeoutMS: parseInt(process.env.MONGODB_CONNECT_TIMEOUT_MS) || 5000,
      serverSelectionTimeoutMS: parseInt(process.env.MONGODB_SERVER_SELECTION_TIMEOUT_MS) || 5000,
    }
  },
  redis: {
    host: process.env.REDIS_HOST || 'localhost',
    port: parseInt(process.env.REDIS_PORT) || 6380,
    password: process.env.REDIS_PASSWORD || '',
    db: parseInt(process.env.REDIS_DB) || 2,
  },
  jwt: {
    secret: process.env.JWT_SECRET || 'a82552a2c8133eddce94cc781f716cdcb911d065528783a8a75256aff6731886',
    algorithm: process.env.JWT_ALGORITHM || 'HS256',
    expiryHours: parseInt(process.env.JWT_EXPIRY_HOURS) || 24,
  },
  cors: {
    origin: [
      // Local development origins
      "http://localhost:5173",      // Vite dev server (default)
      "http://localhost:5174",      // Vite dev server (Docker simulation)
      "http://localhost:5175",      // Electron dev server
      "http://localhost:4173",      // Vite preview (local)
      "http://localhost:4174",      // Vite preview (Docker)
      "http://localhost:3000",      // Legacy frontend port
      "http://localhost:8080",      // Nginx gateway (local)
      "http://localhost:8081",  
      "http://127.0.0.1:5173",      // Alternative localhost
      "http://127.0.0.1:8080",      // Alternative localhost gateway
      // Docker/Production origins
      "http://gateway:80",          // Docker nginx gateway
      "http://gateway:8080",        // Docker nginx gateway alt port
      "http://gateway:8081", 
      "https://gateway:443",        // Docker nginx gateway HTTPS
      // Service-to-service communication
      "http://api:8000",            // Go backend service
      "http://backend-python:8001", // Python backend service
      "http://notes-server:5000",   // Notes server service
      // HTTPS variants
      "https://localhost:443",
      "https://127.0.0.1:443",
      // WebSocket origins (same as HTTP but may be needed separately)
      "ws://localhost:5173",
      "ws://localhost:5174", 
      "ws://localhost:5175",
      "ws://localhost:4173",
      "ws://localhost:4174",
      "ws://localhost:8080",
      "ws://127.0.0.1:5173",
      "ws://127.0.0.1:8080",
      "wss://localhost:443",
      "wss://127.0.0.1:443"
    ],
    credentials: true,
    methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS', 'PATCH'],
    allowedHeaders: [
      'Content-Type', 
      'Authorization', 
      'X-Requested-With',
      'Accept',
      'Origin',
      'X-Organization-ID',
      'x-organization-id',
      'X-User-Id',
      'Cache-Control',
      'Pragma',
      'Accept-Encoding',
      'Content-Encoding',
      // WebSocket specific headers
      'Connection',
      'Upgrade',
      'Sec-WebSocket-Key',
      'Sec-WebSocket-Version',
      'Sec-WebSocket-Protocol',
      'Sec-WebSocket-Extensions'
    ],
    exposedHeaders: [
      'X-RateLimit-Remaining',
      'X-RateLimit-Reset',
      'Content-Length',
      'Content-Type'
    ],
    maxAge: 86400
  },
  logging: {
    level: process.env.LOG_LEVEL || 'info',
    format: process.env.LOG_FORMAT || 'combined'
  }
};

const configureServer = (app) => {
  app.use(cors(config.cors));
  app.use(express.json());
  app.use(requestLogger);

  // Error handling middleware
  app.use((err, req, res, next) => {
    logger.error('Unhandled error', { 
      error: err.stack,
      path: req.path,
      method: req.method,
      ip: req.ip
    });
    
    const status = err.status || (err.message.includes('Token') ? 401 : 500);
    const errorResponse = {
      success: false,
      message: err.message || 'Internal Server Error',
      errors: [{
        message: err.message || 'Internal Server Error',
        code: err.code || 'INTERNAL_ERROR'
      }]
    };
    
    res.status(status).json(errorResponse);
  });

  return app;
};

module.exports = { configureServer, config }; 