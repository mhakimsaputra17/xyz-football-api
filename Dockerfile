# =============================================================================
# Dockerfile — Multi-stage build for xyz-football-api
# =============================================================================

# ---------------------------------------------------------------------------
# Stage 1: Dependencies — download Go modules (cached unless go.mod/go.sum change)
# ---------------------------------------------------------------------------
FROM golang:1.25-alpine AS deps

WORKDIR /src

# Copy only dependency manifests first for layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# ---------------------------------------------------------------------------
# Stage 2: Builder — compile the Go binary
# ---------------------------------------------------------------------------
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Re-use downloaded modules from deps stage
COPY --from=deps /go/pkg /go/pkg

# Copy dependency manifests (needed for build)
COPY go.mod go.sum ./

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/
COPY docs/ docs/

# Build a fully static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/server ./cmd/api

# ---------------------------------------------------------------------------
# Stage 3: Runtime — minimal image with only the binary
# ---------------------------------------------------------------------------
FROM alpine:3.21 AS runtime

# Install ca-certificates (for HTTPS calls) and wget (for HEALTHCHECK)
RUN apk add --no-cache ca-certificates wget

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy binary from builder
COPY --from=builder /app/server /app/server

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose application port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
ENTRYPOINT ["/app/server"]
