#!/bin/bash
set -e

echo "==================================="
echo "Send Me Home - Deployment Script"
echo "==================================="
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo "Error: .env file not found!"
    echo "Please create .env from .env.example and configure your credentials"
    echo ""
    echo "  cp .env.example .env"
    echo "  # Edit .env with your credentials"
    exit 1
fi

# Load environment variables
export $(cat .env | grep -v '^#' | xargs)

# Check deployment type
echo "Select deployment type:"
echo "1. Local (Docker Compose)"
echo "2. Google Cloud Run"
echo "3. Build Docker images only"
echo ""
read -p "Enter choice [1-3]: " choice

case $choice in
    1)
        echo ""
        echo "Deploying locally with Docker Compose..."
        echo ""

        # Check if Docker is running
        if ! docker info > /dev/null 2>&1; then
            echo "Error: Docker is not running. Please start Docker and try again."
            exit 1
        fi

        # Build images
        echo "Building Docker images..."
        docker-compose build

        # Start services
        echo "Starting services..."
        docker-compose up -d

        echo ""
        echo "Deployment complete!"
        echo ""
        echo "Application running at: http://localhost:8080"
        echo ""
        echo "View logs with: docker-compose logs -f"
        echo "Stop with:      docker-compose down"
        ;;

    2)
        echo ""
        echo "Deploying to Google Cloud Run..."
        echo ""

        # Check if gcloud is installed
        if ! command -v gcloud &> /dev/null; then
            echo "Error: gcloud CLI not found. Please install Google Cloud SDK:"
            echo "https://cloud.google.com/sdk/docs/install"
            exit 1
        fi

        # Check if GCP_PROJECT_ID is set
        if [ -z "$GOOGLE_CLOUD_PROJECT" ]; then
            echo "Error: GOOGLE_CLOUD_PROJECT not set in .env"
            exit 1
        fi

        export GCP_PROJECT_ID="$GOOGLE_CLOUD_PROJECT"

        # Set project
        gcloud config set project $GCP_PROJECT_ID

        # Deploy using Make
        make deploy-gcp
        ;;

    3)
        echo ""
        echo "Building Docker images..."
        echo ""

        docker build -t send-me-home:latest .

        echo ""
        echo "Image built successfully!"
        echo ""
        docker images | grep send-me-home
        ;;

    *)
        echo "Invalid choice. Exiting."
        exit 1
        ;;
esac

echo ""
echo "Done!"
