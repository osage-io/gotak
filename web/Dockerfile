# Multi-stage Dockerfile for GoTAK Web UI
# Stage 1: Build the React application
FROM node:20-alpine as build

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies (including dev dependencies for build)
RUN npm ci

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Stage 2: Serve with nginx (unprivileged: non-root, listens on 8080 —
# required for OpenShift's restricted SCC)
FROM nginxinc/nginx-unprivileged:1.25-alpine

# Image setup needs root; the final USER drops back to nginx (101).
USER root

# Install curl for health checks
RUN apk add --no-cache curl

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/

# Copy built application from build stage
COPY --from=build /app/dist /usr/share/nginx/html

# Copy environment configuration script
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Create directory for runtime config (writable by the runtime group — OpenShift
# assigns an arbitrary UID in group 0)
RUN mkdir -p /usr/share/nginx/html/config && \
    chown -R 101:0 /usr/share/nginx/html && \
    chmod -R g+rwX /usr/share/nginx/html

# Drop back to the unprivileged user
USER 101

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Use custom entrypoint
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
