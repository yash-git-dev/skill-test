#!/bin/bash

# Student Report Service Setup Script

set -e

echo "🚀 Setting up Student Report Service..."

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or higher."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
REQUIRED_VERSION="1.21"

if ! printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V -C; then
    echo "❌ Go version $GO_VERSION is too old. Please install Go $REQUIRED_VERSION or higher."
    exit 1
fi

echo "✅ Go version $GO_VERSION is compatible"

# Create necessary directories
echo "📁 Creating directories..."
mkdir -p reports
mkdir -p logs

# Download dependencies
echo "📦 Downloading dependencies..."
go mod tidy

# Build the service
echo "🔨 Building the service..."
go build -o student-report-service cmd/main.go

# Set environment variables if not set
export GO_SERVICE_PORT=${GO_SERVICE_PORT:-8080}
export NODEJS_API_URL=${NODEJS_API_URL:-http://localhost:5007/api/v1}
export LOG_LEVEL=${LOG_LEVEL:-info}
export LOG_FORMAT=${LOG_FORMAT:-json}
export REPORT_OUTPUT_DIR=${REPORT_OUTPUT_DIR:-./reports}

echo "✅ Setup completed successfully!"
echo ""
echo "📋 Configuration:"
echo "   Service Port: $GO_SERVICE_PORT"
echo "   Node.js API: $NODEJS_API_URL"
echo "   Log Level: $LOG_LEVEL"
echo "   Reports Dir: $REPORT_OUTPUT_DIR"
echo ""
echo "🚀 To start the service:"
echo "   ./student-report-service"
echo ""
echo "🔍 To test the service:"
echo "   curl http://localhost:$GO_SERVICE_PORT/health"
echo ""
echo "📖 For more information, see README.md" 