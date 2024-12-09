FROM golang:1.23.4-alpine AS build

WORKDIR /usr/src/app

# Copy go.mod for dependency management
COPY src/go.mod ./
RUN go mod download && go mod verify

# Copy the entire source directory structure (not just files)
COPY src/ ./

# Build the main package only
RUN go build -v -o /usr/local/bin/app .

FROM alpine:3.20.1
COPY --from=build /usr/local/bin/app /usr/local/bin/app

ENTRYPOINT ["/usr/local/bin/app"]

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 CMD wget --spider --quiet http://localhost:8080 || exit 1

# Expose port 8080 to the outside world
EXPOSE 8080