const winston = require('winston');
const path = require('path');
require('winston-daily-rotate-file');

// Define log levels
const levels = {
  error: 0,
  warn: 1,
  info: 2,
  http: 3,
  debug: 4,
};

// Define log level based on environment
const level = () => {
  const env = process.env.NODE_ENV || 'development';
  const logLevel = process.env.LOG_LEVEL || (env === 'development' ? 'debug' : 'warn');
  return logLevel;
};

// Define colors for each level
const colors = {
  error: 'red',
  warn: 'yellow',
  info: 'green',
  http: 'magenta',
  debug: 'blue',
};

// Add colors to winston
winston.addColors(colors);

// Custom format for HTTP requests
const httpFormat = winston.format.printf(({ level, message, timestamp, ...metadata }) => {
  let msg = `${timestamp} [${level}] : ${message}`;
  if (metadata.req) {
    const { method, url, ip, status, responseTime } = metadata.req;
    msg = `${timestamp} [${level}] ${method} ${url} ${status} ${responseTime}ms - ${ip}`;
  }
  if (metadata.error) {
    msg += `\nError: ${metadata.error}`;
  }
  return msg;
});

// Define the format for logs
const format = winston.format.combine(
  winston.format.timestamp({ format: 'YYYY-MM-DD HH:mm:ss:ms' }),
  winston.format.colorize({ all: true }),
  winston.format.metadata({ fillExcept: ['message', 'level', 'timestamp'] }),
  winston.format.printf(({ level, message, timestamp, metadata }) => {
    let msg = `${timestamp} [${level}] : ${message}`;
    if (Object.keys(metadata).length > 0) {
      msg += `\nMetadata: ${JSON.stringify(metadata, null, 2)}`;
    }
    return msg;
  })
);

// Define which transports the logger must use
const transports = [
  // Console transport for all logs
  new winston.transports.Console({
    format: winston.format.combine(
      winston.format.colorize({ all: true }),
      format
    )
  }),
  
  // Daily rotate file transport for error logs
  new winston.transports.DailyRotateFile({
    filename: path.join(process.env.LOG_FILE_PATH || 'logs', 'error-%DATE%.log'),
    datePattern: 'YYYY-MM-DD',
    level: 'error',
    maxSize: process.env.LOG_MAX_SIZE || '20m',
    maxFiles: process.env.LOG_MAX_FILES || '14d',
    format: winston.format.combine(
      winston.format.uncolorize(),
      format
    )
  }),
  
  // Daily rotate file transport for all logs
  new winston.transports.DailyRotateFile({
    filename: path.join(process.env.LOG_FILE_PATH || 'logs', 'combined-%DATE%.log'),
    datePattern: 'YYYY-MM-DD',
    maxSize: process.env.LOG_MAX_SIZE || '20m',
    maxFiles: process.env.LOG_MAX_FILES || '14d',
    format: winston.format.combine(
      winston.format.uncolorize(),
      format
    )
  }),

  // Daily rotate file transport for HTTP logs
  new winston.transports.DailyRotateFile({
    filename: path.join(process.env.LOG_FILE_PATH || 'logs', 'http-%DATE%.log'),
    datePattern: 'YYYY-MM-DD',
    level: 'http',
    maxSize: process.env.LOG_MAX_SIZE || '20m',
    maxFiles: process.env.LOG_MAX_FILES || '14d',
    format: winston.format.combine(
      winston.format.uncolorize(),
      httpFormat
    )
  })
];

// Create the logger instance
const logger = winston.createLogger({
  level: level(),
  levels,
  format,
  transports,
  // Don't exit on error
  exitOnError: false
});

// Create a stream object for Morgan
logger.stream = {
  write: (message) => {
    logger.http(message.trim());
  }
};

// Add request logging middleware
const requestLogger = (req, res, next) => {
  const start = Date.now();
  
  // Log request
  logger.http('Incoming request', {
    req: {
      method: req.method,
      url: req.originalUrl,
      ip: req.ip,
      userAgent: req.get('user-agent')
    }
  });

  // Log response
  res.on('finish', () => {
    const duration = Date.now() - start;
    logger.http('Request completed', {
      req: {
        method: req.method,
        url: req.originalUrl,
        status: res.statusCode,
        responseTime: duration,
        ip: req.ip
      }
    });
  });

  next();
};

module.exports = { logger, requestLogger }; 