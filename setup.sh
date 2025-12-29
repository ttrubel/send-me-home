#!/bin/bash

set -e

echo "🚀 Setting up Send Me Home project..."

# Check prerequisites
echo ""
echo "Checking prerequisites..."

if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+"
    exit 1
fi
echo "✅ Go installed: $(go version)"

if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 18+"
    exit 1
fi
echo "✅ Node.js installed: $(node --version)"

if ! command -v buf &> /dev/null; then
    echo "⚠️  Buf CLI not found. Installing..."
    go install github.com/bufbuild/buf/cmd/buf@latest
    echo "✅ Buf CLI installed"
else
    echo "✅ Buf CLI installed: $(buf --version)"
fi

# Setup backend
echo ""
echo "📦 Setting up backend..."
cd backend
go mod download
echo "✅ Backend dependencies installed"
cd ..

# Setup frontend
echo ""
echo "📦 Setting up frontend..."
cd frontend
npm install
echo "✅ Frontend dependencies installed"
cd ..

# Generate code from proto
echo ""
echo "🔧 Generating code from proto definitions..."
buf generate
echo "✅ Code generation complete"

echo ""
echo "✨ Setup complete!"
echo ""
echo "To start development:"
echo "  1. Terminal 1: cd backend && go run cmd/server/main.go"
echo "  2. Terminal 2: cd frontend && npm run dev"
echo ""
echo "Or use Make commands:"
echo "  make dev-backend"
echo "  make dev-frontend"
