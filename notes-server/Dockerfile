# 1. Build stage
FROM node:20-alpine AS builder
WORKDIR /app

# Copy package files and install dependencies
COPY package*.json ./
RUN npm ci

# Copy the rest of the application source code
COPY . .

# (Optional) Add a build step when set it
# RUN npm run build

# 2. Production stage
FROM node:20-alpine
WORKDIR /app

# Install curl for health checks
RUN apk add --no-cache curl

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Create logs directory and set permissions BEFORE switching to non-root user
RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app

# Switch to the non-root user
USER appuser

# Copy dependencies from the builder stage
COPY --from=builder /app/node_modules ./node_modules
# Copy application code from builder stage, ensuring correct ownership
COPY --from=builder --chown=appuser:appgroup /app .

# Set environment variables for production
ENV NODE_ENV=production
ENV DOCKER_ENV=true

EXPOSE 5000

# Start the server
CMD ["node", "cmd/api/server.js"] 