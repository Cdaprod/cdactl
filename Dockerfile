# Dockerfile for building and running cdactl with TUI

# Use an official Go runtime as a base image
FROM golang:1.20-alpine as build

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files for dependencies
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the rest of the application's source code
COPY . .

# Build the Go binary
RUN go build -o /cdactl ./cmd/cdactl/main.go

# Create a smaller runtime image
FROM alpine:latest

# Copy the compiled binary from the build stage
COPY --from=build /cdactl /usr/local/bin/cdactl

# Set the command to execute the TUI on container start
CMD ["cdactl"]