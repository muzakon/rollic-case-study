#!/bin/sh
set -e

echo "Running migrations..."
go run cmd/migrate/main.go up

echo "Starting server with live reload..."
exec air -c .air.toml
