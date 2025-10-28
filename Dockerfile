FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install ca-certificates and tzdata in builder stage
RUN apk --no-cache add ca-certificates tzdata

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO disabled for static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags='-w -s -extldflags "-static"' -o exchange-server ./cmd/server

# Final stage using distroless
FROM gcr.io/distroless/static-debian12:nonroot

# Copy ca-certificates and timezone data from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary from builder stage
COPY --from=builder /app/exchange-server /exchange-server

# Copy default config and production config
COPY --from=builder /app/config.yaml /config.yaml
COPY --from=builder /app/config.production.yaml /config.production.yaml

# Use ARG for configurable port
ARG PORT=80
ENV PORT=${PORT}

# Environment variable for config selection
ENV CONFIG_FILE=/config.yaml

# Expose the configurable port
EXPOSE ${PORT}

# Command to run with configurable config file
CMD ["/exchange-server"]