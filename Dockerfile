# Dockerfile for cdactl

# Use an official Go runtime as a build stage
FROM golang:1.20-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project into the container
COPY . .

# Build the Go binary
RUN go build -o /cdactl ./cmd/cdactl/main.go

# Use a minimal base image for the runtime
FROM alpine:latest

# Set up the config directory
RUN mkdir -p /root/.config/cdactl/

# Copy the compiled binary and config.yaml
COPY --from=builder /cdactl /usr/local/bin/cdactl
COPY .config/cdactl/config.yaml /root/.config/cdactl/config.yaml

# Set the entry point to cdactl
ENTRYPOINT ["cdactl"]

# Default command to show help
CMD ["help"]