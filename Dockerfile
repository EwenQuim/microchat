# syntax=docker/dockerfile:1
# Build stage for frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY app/package*.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci
COPY app/ ./
RUN --mount=type=cache,target=/root/.npm \
    npm run build

# Build stage for Go
FROM golang:1.25-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY . .
# Copy frontend build to be embedded
COPY --from=frontend-builder /app/frontend/dist ./cmd/server/static
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o cli ./cmd/cli

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy Go binaries (static files are now embedded in the binary)
COPY --from=go-builder /app/server .
COPY --from=go-builder /app/cli .

# Run server
CMD ["./server"]
