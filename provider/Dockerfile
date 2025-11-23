FROM golang:1.25-alpine

# Install air for hot reloading and other dev tools
RUN go install github.com/air-verse/air@latest

# Install build tools
RUN apk add --no-cache git gcc musl-dev

RUN apk add --no-cache bash git libstdc++ libc6-compat

WORKDIR /app

# Copy go mod files first (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code (will be overridden by volume mount)
COPY . .

EXPOSE 8080

# Air will watch for changes and rebuild
CMD ["air", "-c", ".air.toml"]