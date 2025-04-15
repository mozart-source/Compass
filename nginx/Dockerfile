# Stage 1: Build the frontend assets
FROM node:19.0.1-alpine AS frontend-builder
WORKDIR /app

# First, copy the .dockerignore file and then the package files
COPY Frontend/.dockerignore ./.dockerignore
COPY Frontend/package*.json ./

# Install dependencies
RUN npm ci --legacy-peer-deps

# Copy the rest of the frontend source code
COPY Frontend/ ./

# Build the frontend with Docker environment variables
ENV NODE_ENV=production
ENV DOCKER_ENV=true
ENV VITE_API_URL=/api
ENV VITE_PYTHON_API_URL=/api/v1
ENV VITE_NOTES_URL=/notes
# Debug the build process
RUN echo "Building frontend with NODE_ENV=$NODE_ENV and DOCKER_ENV=$DOCKER_ENV"
RUN npm run build:docker

# Stage 2: Set up the Nginx server
FROM nginx:1.25-alpine

# Copy the Nginx configuration
COPY nginx/nginx.conf /etc/nginx/nginx.conf

# Copy test HTML file
COPY nginx/test.html /usr/share/nginx/html/test.html

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy the built frontend assets from the builder stage
COPY --from=frontend-builder /app/dist /usr/share/nginx/html

# Copy SSL certificates
COPY certs /etc/nginx/certs

# Create directories for PID and logs, and set permissions
RUN mkdir -p /var/run/nginx && \
    mkdir -p /var/log/nginx && \
    mkdir -p /var/cache/nginx && \
    chown -R nginx:nginx /var/run/nginx && \
    chown -R nginx:nginx /var/log/nginx && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /usr/share/nginx/html

# RUN nginx -t


EXPOSE 80
EXPOSE 443

# Start nginx with explicit configuration file
CMD ["nginx", "-g", "daemon off;", "-c", "/etc/nginx/nginx.conf"] 