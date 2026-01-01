FROM rust:1-alpine

# Install cargo-watch for hot reloading and build dependencies
RUN cargo install cargo-watch

# Install build tools and utilities
RUN apk add --no-cache git gcc musl-dev make bash libstdc++ libc6-compat

WORKDIR /app

# Copy source code (will be overridden by volume mount in development)
# Note: Cargo.toml should exist for the build to work properly
COPY . .

# Default command (can be overridden in docker-compose)
CMD ["cargo", "watch", "-x", "run"]

