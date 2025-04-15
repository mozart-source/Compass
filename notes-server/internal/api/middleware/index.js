const { requestLogger } = require('../../../pkg/utils/logger');
const createRateLimiter = require('./rateLimiter');
const { extractUserIdFromToken } = require('../../../pkg/utils/jwt');

function userContextMiddleware(req, res, next) {
  req.user = null;
  if (req.headers.authorization) {
    try {
      const userId = extractUserIdFromToken(req.headers.authorization);
      if (userId) {
        req.user = { id: userId };
      }
    } catch (e) {
      return next(e);
    }
  }
  next();
}

module.exports = {
  requestLogger,
  createRateLimiter,
  userContextMiddleware,
};