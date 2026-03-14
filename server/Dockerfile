FROM golang:1.25.5

# Set working directory
WORKDIR /app

# Install air for hot reload
RUN go install github.com/air-verse/air@latest

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Run air for hot reload, fallback to regular build
CMD ["sh", "-c", "air -c .air.toml || air --build.cmd 'go build -o ./tmp/main ./cmd/api' --build.bin './tmp/main'"]
