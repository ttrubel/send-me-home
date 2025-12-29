.PHONY: help install dev build test clean docker-build deploy

help:
	@echo "Send Me Home - Development Commands"
	@echo ""
	@echo "install        Install all dependencies (backend + frontend)"
	@echo "dev            Run both backend and frontend in development mode"
	@echo "build          Build both backend and frontend"
	@echo "test           Run all tests"
	@echo "clean          Clean build artifacts"
	@echo "docker-build   Build Docker image for backend"
	@echo "deploy         Deploy to Google Cloud"

install:
	@echo "Installing backend dependencies..."
	cd backend && go mod download
	@echo "Installing frontend dependencies..."
	cd frontend && npm install

dev:
	@echo "Starting development servers..."
	@echo "Run 'make dev-backend' and 'make dev-frontend' in separate terminals"

dev-backend:
	cd backend && go run cmd/server/main.go

dev-frontend:
	cd frontend && npm start

build:
	@echo "Building backend..."
	cd backend && go build -o bin/server cmd/server/main.go
	@echo "Building frontend..."
	cd frontend && npm run build

test:
	@echo "Running backend tests..."
	cd backend && go test ./...
	@echo "Running frontend tests..."
	cd frontend && npm test -- --watchAll=false

clean:
	@echo "Cleaning build artifacts..."
	rm -rf backend/bin backend/tmp
	rm -rf frontend/build frontend/dist
	@echo "Clean complete"

docker-build:
	cd backend && docker build -t send-me-home-backend .

deploy:
	@echo "Deploying to Google Cloud..."
	@echo "TODO: Add deployment scripts"
