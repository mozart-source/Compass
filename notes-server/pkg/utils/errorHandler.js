const { z } = require('zod');
const { logger } = require('./logger');

class BaseError extends Error {
  constructor(message, code = 'INTERNAL_ERROR') {
    super(message);
    this.name = this.constructor.name;
    this.code = code;
    this.timestamp = new Date().toISOString();
    
    // Log the error
    logger.error(message, {
      error: {
        name: this.name,
        code: this.code,
        stack: this.stack
      }
    });
  }
}

class ValidationError extends BaseError {
  constructor(message, field = null) {
    super(message, 'VALIDATION_ERROR');
    this.field = field;
  }
}

class NotFoundError extends BaseError {
  constructor(entity) {
    super(`${entity} not found`, 'NOT_FOUND');
    this.entity = entity;
  }
}

class DatabaseError extends BaseError {
  constructor(message) {
    super(message, 'DATABASE_ERROR');
  }
}

class AuthenticationError extends BaseError {
  constructor(message) {
    super(message, 'AUTHENTICATION_ERROR');
  }
}

class AuthorizationError extends BaseError {
  constructor(message) {
    super(message, 'AUTHORIZATION_ERROR');
  }
}

const handleZodError = (error) => {
  if (!(error instanceof z.ZodError)) {
    throw error;
  }

  logger.warn('Zod Validation Error', { errors: error.errors });
  return error.errors.map(err => ({
    message: err.message,
    field: err.path.join('.'),
    code: 'VALIDATION_ERROR'
  }));
};

const formatGraphQLError = (error) => {
  const originalError = error.originalError;

  // Surface GraphQL validation errors (e.g., unknown argument/field)
  if (error.message && error.message.match(/Unknown argument|Cannot query field/)) {
    // Try to extract the field/argument name from the message
    const fieldMatch = error.message.match(/"([^"]+)"/);
    return {
      message: error.message,
      code: 'GRAPHQL_VALIDATION_ERROR',
      field: fieldMatch ? fieldMatch[1] : undefined,
      timestamp: new Date().toISOString()
    };
  }

  if (originalError instanceof BaseError) {
    return {
      message: originalError.message,
      code: originalError.code,
      field: originalError.field,
      timestamp: originalError.timestamp
    };
  }

  // Log unexpected errors
  logger.error('Unexpected GraphQL error', {
    error: {
      message: error.message,
      stack: error.stack,
      originalError: originalError?.message
    }
  });

  return {
    message: error.message || 'An unexpected error occurred',
    code: 'INTERNAL_ERROR',
    field: undefined,
    timestamp: new Date().toISOString()
  };
};

module.exports = {
  BaseError,
  ValidationError,
  NotFoundError,
  DatabaseError,
  AuthenticationError,
  AuthorizationError,
  handleZodError,
  formatGraphQLError
}; 