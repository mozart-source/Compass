const jwt = require('jsonwebtoken');
const { config } = require('../../internal/config/server');

function extractUserIdFromToken(authorization) {
  if (!authorization || !authorization.startsWith('Bearer ')) {
    throw new Error('Invalid or missing token');
  }
  const token = authorization.split(' ')[1];
  try {
    const claims = jwt.verify(token, config.jwt.secret, {
      algorithms: [config.jwt.algorithm]
    });
    
    // Handle multiple user ID field formats for cross-service compatibility
    let userId = claims.user_id || claims.userId || claims.sub;
    
    if (!userId) {
      throw new Error('user_id not found in token');
    }
    
    // Convert to string if it's a UUID object or other format
    if (typeof userId === 'object' && userId !== null) {
      userId = String(userId);
    } else if (typeof userId !== 'string') {
      userId = String(userId);
    }
    
    return userId;
  } catch (e) {
    throw new Error(`Token decode error: ${e.message}`);
  }
}

module.exports = { extractUserIdFromToken }; 