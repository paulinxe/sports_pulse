FROM golang:1.26-alpine

# Install air for hot reloading and other dev tools
RUN go install github.com/air-verse/air@latest

# Install build tools
RUN apk add --no-cache git gcc musl-dev

RUN apk add --no-cache git gcc musl-dev make bash libstdc++ libc6-compat

# Ensure Go is in PATH (must be set early)
ENV PATH="/usr/local/go/bin:${PATH}"

# Add Go to PATH in profile.d for all shells (Alpine Linux standard)
RUN echo 'export PATH="/usr/local/go/bin:$PATH"' > /etc/profile.d/go.sh && \
    chmod +x /etc/profile.d/go.sh

# Install golangci-lint with same Go as image (so version matches go.mod)
ENV PATH="/go/bin:${PATH}"
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

WORKDIR /app

# Copy go mod files first (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code (will be overridden by volume mount)
COPY . .

# Create user matching host IDs
ARG UID=1000 # Default value if not provided
ARG GID=1000 # Default value if not provided
RUN addgroup -g ${GID} appgroup && \
    adduser -u ${UID} -G appgroup -D -g "" appuser

RUN chown appuser:appgroup /app

# Ensure appuser can write to .cache (golangci-lint/goimports, gopls, etc.)
RUN mkdir -p /home/appuser/.cache && chown -R appuser:appgroup /home/appuser

ENV GOCACHE=/home/appuser/.cache/go-cache

### Development only
RUN go install golang.org/x/tools/gopls@latest
RUN chown -R appuser:appgroup /home/appuser/.cache/go-cache
### End of development only

USER appuser

EXPOSE 8080

# Air will watch for changes and rebuild
CMD ["air", "-c", ".air.toml"]
