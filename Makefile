.PHONY: help install dev build test clean docker-build docker-up docker-down deploy

help:
	@echo "Send Me Home - Development Commands"
	@echo ""
	@echo "install          Install all dependencies (backend + frontend)"
	@echo "dev              Run both backend and frontend in development mode"
	@echo "build            Build both backend and frontend"
	@echo "test             Run all tests"
	@echo "clean            Clean build artifacts"
	@echo "docker-build     Build Docker images for backend and frontend"
	@echo "docker-up        Start all services with Docker Compose"
	@echo "docker-down      Stop all services with Docker Compose"
	@echo "deploy-gcp       Deploy to Google Cloud Run"

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
	@echo "Building unified Docker image..."
	docker build -t send-me-home:latest .
	@echo "Docker image built successfully"

docker-up:
	@echo "Starting service with Docker Compose..."
	docker-compose up -d
	@echo "Application started at http://localhost:8080"

docker-down:
	@echo "Stopping Docker Compose services..."
	docker-compose down
	@echo "Services stopped"

deploy-gcp:
	@echo "Deploying to Google Cloud Run..."
	@echo "Make sure you have set GCP_PROJECT_ID environment variable"
	@if [ -z "$$GCP_PROJECT_ID" ]; then \
		echo "Error: GCP_PROJECT_ID not set"; \
		exit 1; \
	fi
	@echo "Building and pushing unified image..."
	gcloud builds submit --tag gcr.io/$$GCP_PROJECT_ID/send-me-home
	@echo "Deploying to Cloud Run..."
	gcloud run deploy send-me-home \
		--image gcr.io/$$GCP_PROJECT_ID/send-me-home \
		--platform managed \
		--region us-central1 \
		--allow-unauthenticated \
		--set-env-vars GOOGLE_GENAI_USE_VERTEXAI=true,GOOGLE_CLOUD_PROJECT=$$GCP_PROJECT_ID
	@echo "Getting service URL..."
	$(eval APP_URL := $(shell gcloud run services describe send-me-home --platform managed --region us-central1 --format 'value(status.url)'))
	@echo ""
	@echo "Deployment complete!"
	@echo "Application URL: $(APP_URL)"
