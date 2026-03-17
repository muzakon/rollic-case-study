FROM golang:1.25

WORKDIR /app

# Install air for live reload
RUN go install github.com/air-verse/air@latest

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download