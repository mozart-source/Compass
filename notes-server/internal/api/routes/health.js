const express = require('express');
const router = express.Router();
const mongoose = require('mongoose');
const { logger } = require('../../../pkg/utils/logger');

router.get('/', (req, res) => {
  const dbStatus = mongoose.connection.readyState === 1 ? 'connected' : 'disconnected';
  const redisStatus = global.redisClient?.isHealthy() ? 'connected' : 'disconnected';
  
  logger.info('Health check', { 
    database: dbStatus,
    redis: redisStatus
  });
  
  res.json({
    status: 'ok',
    timestamp: new Date().toISOString(),
    database: dbStatus,
    redis: redisStatus
  });
});

module.exports = router; 