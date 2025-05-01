# Stage 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Set Go environment variables for static build
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Build the binary using the Go build cache.
RUN --mount=type=cache,id=go-build,target=/root/.cache/go-build \
    go build -v -o pocketbase_amd64

# Stage 2: Final image
FROM alpine:3.16 AS runtime

WORKDIR /

# Copy the binary from the builder stage
COPY --from=builder /app/pocketbase_amd64 /pocketbase

# Copy migrations
COPY migrations ./migrations

# Make sure the binary is executable
RUN chmod +x /pocketbase

# Verify binary output contains expected version string
RUN output="$(/pocketbase --version)" && \
    echo "Version output: $output" && \
    echo "$output" | grep -q 'pocketbase version (untracked)'

ENV MEILI_URL=http://localhost:7700
ENV MEILI_MASTER_KEY=

EXPOSE 8090

ENTRYPOINT ["/pocketbase", "serve", "--http=0.0.0.0:8090", "--dir=/pb_data"]