FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install ca-certificates, tzdata, git, and bash in builder stage
RUN apk --no-cache add ca-certificates tzdata git bash

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Inject version information into web files
ARG BUILD_DATE
ARG GIT_COMMIT
RUN if [ -f scripts/inject-version.sh ]; then \
        echo "Injecting version info..." && \
        COMMIT_SHORT=$(echo ${GIT_COMMIT:-unknown} | cut -c1-7) && \
        BUILD_DATE=${BUILD_DATE:-$(date -u +"%Y-%m-%d %H:%M:%S UTC")} && \
        sed -i "s/{{VERSION}}/$COMMIT_SHORT/g" web/index.html && \
        sed -i "s/{{COMMIT}}/${GIT_COMMIT:-unknown}/g" web/index.html && \
        sed -i "s/{{BUILD_DATE}}/$BUILD_DATE/g" web/index.html && \
        echo "✅ Version $COMMIT_SHORT injected"; \
    fi

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

# Copy web directory for static files
COPY --from=builder /app/web /web

# Use standard HTTP port for Cloudflare compatibility
ENV PORT=80

# Environment variable for config selection
ENV CONFIG_FILE=/config.yaml

# Expose standard HTTP port (Cloudflare compatible)
EXPOSE 80

# Command to run with configurable config file
CMD ["/exchange-server"]