// Enhanced environment detection for frontend configuration
export const GO_API_URL = '/api';
export const PYTHON_API_URL = '/api/v1';
export const NOTES_API_URL = '/notes';

// Enhanced environment detection function
const detectEnvironment = () => {
  const isDevelopment = process.env.NODE_ENV === 'development';
  const isLocalhost = window.location.hostname === 'localhost' || 
                      window.location.hostname === '127.0.0.1' ||
                      window.location.hostname === '0.0.0.0';
  
  // Check if running through Vite dev server (updated ports)
  const isViteDevServer = window.location.port === '5173' || 
                          window.location.port === '5174' || 
                          window.location.port === '5175' || 
                          window.location.port === '4173' || 
                          window.location.port === '4174';
  
  // Check if running through nginx gateway (port 8081 in Docker setup)
  const isNginxGateway = window.location.port === '8081' || 
                         window.location.port === '80' || 
                         window.location.port === '443' ||
                         window.location.port === '';
  
  // Docker environment indicators - if not localhost OR running through gateway
  const isDocker = !isLocalhost || isNginxGateway || process.env.DOCKER_ENV === 'true';
  
  // Check if current page is HTTPS
  const isHttps = window.location.protocol === 'https:';
  
  const envInfo = {
    isDevelopment,
    isLocalhost,
    isViteDevServer,
    isNginxGateway,
    isDocker,
    isHttps,
    hostname: window.location.hostname,
    port: window.location.port,
    origin: window.location.origin
  };
  
  console.log('Frontend Environment Detection:', envInfo);
  
  return envInfo;
};

// Main configuration function
export const getApiUrls = () => {
  const env = detectEnvironment();
  
  console.log('=== API URL Configuration ===');
  console.log('Environment Info:', env);
  
  // LOCAL DEVELOPMENT: Direct service URLs
  if (env.isDevelopment && env.isLocalhost && (env.isViteDevServer || !env.isNginxGateway)) {
    console.log('Using LOCAL DEVELOPMENT URLs');
    
    // Use WSS for HTTPS pages, WS for HTTP pages
    const wsProtocol = env.isHttps ? 'wss' : 'ws';
    
    const urls = {
      GO_API_URL: 'http://localhost:8000/api',
      PYTHON_API_URL: 'http://localhost:8001/api/v1', 
      NOTES_API_URL: 'http://localhost:5000',
      WS_BASE_URL: `${wsProtocol}://localhost`,
      // Individual WebSocket URLs for different services
      WS_GO_URL: `${wsProtocol}://localhost:8000`,
      WS_PYTHON_URL: `${wsProtocol}://localhost:8001/api/v1`,
      WS_NOTES_URL: `${wsProtocol}://localhost:5000`,
    };
    
    console.log('Generated LOCAL URLs:', urls);
    console.log('============================');
    return urls;
  }
  
  // DOCKER/PRODUCTION: Nginx gateway with relative URLs
  console.log('Using DOCKER/PRODUCTION URLs (via nginx gateway)');
  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsHost = window.location.host;
  
  const urls = {
    GO_API_URL: '/api',
    PYTHON_API_URL: '/api/v1',
    NOTES_API_URL: '/notes',
    WS_BASE_URL: `${wsProtocol}//${wsHost}`,
    // Individual WebSocket URLs for different services (all through nginx gateway)
    WS_GO_URL: `${wsProtocol}//${wsHost}`,
    WS_PYTHON_URL: `${wsProtocol}//${wsHost}/api/v1`,
    WS_NOTES_URL: `${wsProtocol}//${wsHost}`,
  };
  
  console.log('Generated DOCKER URLs:', urls);
  console.log('============================');
  return urls;
};

// Utility function to get the current environment info (for debugging)
export const getEnvironmentInfo = () => {
  return detectEnvironment();
};
