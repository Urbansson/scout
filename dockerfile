# Builder stage
FROM golang:1.24 AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY *.go ./

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scout .

# Runner stage - use distroless
FROM gcr.io/distroless/static-debian12:nonroot

# OCI labels
LABEL org.opencontainers.image.description="Discord bot that retrieves and reports the server's public IP address"

# Copy binary from builder
COPY --from=builder /app/scout /scout

# Use non-root user
USER nonroot:nonroot

# Run
ENTRYPOINT ["/scout"]
