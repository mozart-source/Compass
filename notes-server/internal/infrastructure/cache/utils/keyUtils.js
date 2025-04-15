function paramStringify(params = {}) {
  return Object.entries(params)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([k, v]) => `${k}:${JSON.stringify(v)}`)
    .join('|');
}

function generateEntityKey(keyPrefix, entityType, entityId, action = '') {
  return `${keyPrefix}${entityType}:${entityId}${action ? ':' + action : ''}`;
}

function generateListKey(userId, entityType, params = {}) {
  return `user:${userId}:${entityType}:${paramStringify(params)}`;
}

function generateTagSetKey(keyPrefix, tag) {
  return `${keyPrefix}tag:${tag}`;
}

module.exports = {
  generateEntityKey,
  generateListKey,
  generateTagSetKey,
  paramStringify
}; 