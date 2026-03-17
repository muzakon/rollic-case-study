FROM golang:1.25

WORKDIR /app

# Install air for live reload and gotestsum for testing
RUN go install github.com/air-verse/air@latest
RUN go install gotest.tools/gotestsum@latest

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download