FROM rust:1-alpine

# Ensure cargo is in PATH (must be set early)
ENV PATH="/usr/local/cargo/bin:${PATH}"

# Add cargo to PATH in profile.d for all shells (Alpine Linux standard)
# For development purposes
RUN echo 'export PATH="/usr/local/cargo/bin:$PATH"' > /etc/profile.d/cargo.sh && \
    chmod +x /etc/profile.d/cargo.sh

# Install rustfmt for code formatting
RUN rustup component add rustfmt

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

