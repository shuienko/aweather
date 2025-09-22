FROM golang:1.23.4-alpine AS build

WORKDIR /usr/src/app

# Copy go.mod and go.sum for dependency management
COPY src/go.mod src/go.sum ./
RUN go mod download && go mod verify

# Copy the entire source directory structure (not just files)
COPY src/ ./

# Build the main package only
RUN go build -v -o /usr/local/bin/app .

# Install tzdata
RUN apk add --no-cache tzdata

FROM alpine:latest

# Copy application binary
COPY --from=build /usr/local/bin/app /usr/local/bin/app

# Copy tzdata
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

# Install CA certificates for outbound HTTPS (Open‑Meteo, Geocoding)
RUN apk add --no-cache ca-certificates && update-ca-certificates

# Create non-root user and switch
RUN adduser -D -H -s /sbin/nologin appuser
USER appuser

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/app"]

# Application must listen on port 8080 
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 CMD wget --spider --quiet http://localhost:8080 || exit 1

# Expose port 8080 to the outside world
EXPOSE 8080