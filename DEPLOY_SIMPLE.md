# Simple Deployment Guide

This project uses a **single unified Docker image** that contains both the backend API and frontend static files. The Go backend serves the React frontend from the `/public` directory.

## Architecture

```
Single Container:
┌─────────────────────────────┐
│  Go Backend (Port 8080)     │
│  ├─ API endpoints (/game/*) │
│  └─ Static files (/)        │
│     └─ React frontend       │
└─────────────────────────────┘
```

## Quick Start

### Option 1: Docker Compose (Easiest)

```bash
# 1. Configure environment
cp .env.example .env
# Edit .env with your API keys

# 2. Start
make docker-up

# 3. Access
# Open http://localhost:8080
```

### Option 2: Interactive Script

```bash
./deploy.sh
# Select option 1 for local deployment
```

### Option 3: Manual Docker Build

```bash
# Build image
docker build -t send-me-home:latest .

# Run container
docker run -p 8080:8080 \
  -e GOOGLE_GENAI_USE_VERTEXAI=true \
  -e GOOGLE_CLOUD_PROJECT=your-project \
  -e ELEVENLABS_API_KEY=your-key \
  send-me-home:latest

# Access at http://localhost:8080
```

## Production Deployment (Google Cloud Run)

```bash
# Set your project
export GCP_PROJECT_ID=your-project-id

# Deploy
make deploy-gcp

# This will:
# 1. Build the unified image
# 2. Push to Google Container Registry
# 3. Deploy to Cloud Run
# 4. Return the public URL
```

## How It Works

### Build Process (Multi-stage Dockerfile)

```dockerfile
Stage 1: Build Frontend
  - npm install
  - npm run build → creates dist/

Stage 2: Build Backend
  - go mod download
  - go build → creates binary

Stage 3: Runtime
  - Copy backend binary
  - Copy frontend dist/ → /app/public
  - Start backend (serves both API + frontend)
```

### Backend Static File Serving

The Go server in [backend/cmd/server/main.go](backend/cmd/server/main.go:62-88) includes:

```go
// Serve frontend static files from ./public
- Serves static files for / (HTML, JS, CSS, images)
- Serves API at /game.v1.GameService/*
- SPA routing: returns index.html for non-existent files
```

## Environment Variables

```bash
# Required for production
GOOGLE_GENAI_USE_VERTEXAI=true
GOOGLE_CLOUD_PROJECT=your-gcp-project-id
ELEVENLABS_API_KEY=your-elevenlabs-key

# Optional
FIRESTORE_PROJECT_ID=your-firestore-project
PORT=8080 (default)

# For development (skip API calls)
GOOGLE_GENAI_USE_VERTEXAI=false
```

## Benefits of Single Image

✅ **Simpler deployment** - One container instead of two
✅ **No CORS issues** - Same origin for API and frontend
✅ **Easier scaling** - Single service to scale
✅ **Lower costs** - One Cloud Run service instead of two
✅ **Faster builds** - Shared base image layers
✅ **Simpler routing** - No need for reverse proxy

## Development vs Production

### Development (separate servers)
```bash
# Terminal 1: Backend on :8080
make dev-backend

# Terminal 2: Frontend on :5173 (Vite dev server)
make dev-frontend
```

Benefits: Hot reload, faster iteration

### Production (single container)
```bash
make docker-up
```

Benefits: Matches production environment

## File Structure

```
/
├── Dockerfile              # Single unified build
├── docker-compose.yml      # Single service
├── backend/
│   ├── Dockerfile         # (Deprecated - use root)
│   └── cmd/server/main.go # Serves /public
├── frontend/
│   ├── Dockerfile         # (Deprecated - use root)
│   └── dist/              # → Copied to /app/public
└── .dockerignore          # Excludes dev files
```

## Troubleshooting

### Issue: 404 on frontend routes

**Symptom:** Direct navigation to `/game` works, but refresh gives 404

**Fix:** The backend already handles SPA routing. If you still see 404s, check:
1. Frontend build created `dist/index.html`
2. Backend is serving from correct `./public` path

### Issue: API calls fail with CORS

**This shouldn't happen** with unified image since API and frontend are same origin.

If you see CORS errors, you're likely running dev mode with separate servers.

### Issue: Frontend shows old version

**Fix:** Rebuild the image
```bash
docker-compose down
docker-compose build --no-cache
docker-compose up
```

## Commands Reference

```bash
# Development
make dev-backend           # Run backend only
make dev-frontend          # Run frontend only (Vite)

# Docker (unified image)
make docker-build          # Build single image
make docker-up             # Start container
make docker-down           # Stop container

# Production
make deploy-gcp            # Deploy to Cloud Run
./deploy.sh               # Interactive deployment

# Testing
make test                  # Run all tests
make clean                 # Clean build artifacts
```

## Cost Estimate (Google Cloud Run)

**Single unified service:**
- CPU: 1 vCPU
- Memory: 512 MB
- Requests: 100k/month
- **Cost: ~$5-10/month**

Compare to separate services (2x containers): ~$10-20/month

## Next Steps

1. Configure `.env` with your credentials
2. Test locally with `make docker-up`
3. Deploy to Cloud Run with `make deploy-gcp`
4. Set up custom domain (optional)
5. Enable Cloud CDN for static assets (optional)
