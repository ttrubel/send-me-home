# Deployment Guide

This guide covers deployment options for **Send Me Home** game.

## Quick Start with Docker Compose

### Prerequisites
- Docker and Docker Compose installed
- Google Cloud credentials (for Vertex AI)
- ElevenLabs API key

### Step 1: Configure Environment

1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Edit `.env` and fill in your credentials:
```bash
# Required for Vertex AI
GOOGLE_GENAI_USE_VERTEXAI=true
GOOGLE_CLOUD_PROJECT=your-gcp-project-id
GCP_CREDENTIALS_PATH=./gcp-key.json

# Required for voice generation
ELEVENLABS_API_KEY=your-elevenlabs-api-key
```

3. Place your GCP service account key file in the project root as `gcp-key.json`

### Step 2: Build and Run

```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

The application will be available at:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080

### Step 3: Health Check

Check if services are running:
```bash
# Backend health
curl http://localhost:8080/health

# Frontend (should return HTML)
curl http://localhost:3000
```

## Production Deployment Options

### Option 1: Google Cloud Run (Recommended)

**Backend deployment:**
```bash
cd backend

# Build and push to GCR
gcloud builds submit --tag gcr.io/YOUR_PROJECT_ID/send-me-home-backend

# Deploy to Cloud Run
gcloud run deploy send-me-home-backend \
  --image gcr.io/YOUR_PROJECT_ID/send-me-home-backend \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars GOOGLE_GENAI_USE_VERTEXAI=true \
  --set-env-vars GOOGLE_CLOUD_PROJECT=YOUR_PROJECT_ID \
  --set-env-vars ELEVENLABS_API_KEY=your-key
```

**Frontend deployment:**
```bash
cd frontend

# Update .env.production with Cloud Run backend URL
echo "VITE_API_URL=https://your-backend-url.run.app" > .env.production

# Build and push to GCR
gcloud builds submit --tag gcr.io/YOUR_PROJECT_ID/send-me-home-frontend

# Deploy to Cloud Run
gcloud run deploy send-me-home-frontend \
  --image gcr.io/YOUR_PROJECT_ID/send-me-home-frontend \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

### Option 2: AWS ECS/Fargate

1. Push images to ECR:
```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin YOUR_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com

# Tag and push backend
docker tag send-me-home-backend:latest YOUR_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/send-me-home-backend:latest
docker push YOUR_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/send-me-home-backend:latest

# Tag and push frontend
docker tag send-me-home-frontend:latest YOUR_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/send-me-home-frontend:latest
docker push YOUR_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/send-me-home-frontend:latest
```

2. Create ECS task definitions and services via AWS Console or CLI

### Option 3: Kubernetes (GKE, EKS, AKS)

See [kubernetes/](./kubernetes/) directory for example manifests.

## Environment Variables Reference

### Backend

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `PORT` | No | Server port | `8080` |
| `GOOGLE_GENAI_USE_VERTEXAI` | Yes* | Enable Vertex AI | `true` |
| `GOOGLE_CLOUD_PROJECT` | Yes* | GCP project ID | `my-project-123` |
| `GOOGLE_APPLICATION_CREDENTIALS` | Yes* | Path to GCP key | `/app/gcp-key.json` |
| `ELEVENLABS_API_KEY` | Yes | ElevenLabs API key | `sk_xxx...` |
| `FIRESTORE_PROJECT_ID` | No | Firestore project | `my-project-123` |

*Required if using Vertex AI (recommended for production)

### Frontend

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `VITE_API_URL` | Yes | Backend API URL | `http://localhost:8080` |

## Cost Optimization

### Vertex AI (Gemini)
- Uses `gemini-1.5-flash-002` model (cost-effective)
- ~15 API calls per game session
- Estimated cost: $0.01-0.02 per game

### ElevenLabs
- ~15-20 TTS calls per game session
- Uses standard voices (not cloned)
- Estimated cost: $0.05-0.10 per game (depending on plan)

### Tips to Reduce Costs
1. **Use mock mode for development**:
   - Set `GOOGLE_GENAI_USE_VERTEXAI=false` to use mock data
   - No API calls, no costs

2. **Implement caching**:
   - Cache generated rules daily
   - Reuse case data for testing

3. **Use Cloud Run's free tier**:
   - 2M requests/month free
   - Perfect for demos and small traffic

## Monitoring

### Logs

**Docker Compose:**
```bash
docker-compose logs -f backend
docker-compose logs -f frontend
```

**Cloud Run:**
```bash
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=send-me-home-backend" --limit 50
```

### Metrics

Monitor these for production:
- API response times
- Vertex AI API errors
- ElevenLabs API errors
- Frontend 404s
- Memory usage

## Troubleshooting

### Issue: Backend returns mock data in production

**Check:**
```bash
# Verify environment variable
docker exec send-me-home-backend env | grep GOOGLE_GENAI_USE_VERTEXAI

# Check logs for initialization
docker-compose logs backend | grep "Gemini client"
```

**Fix:** Ensure `GOOGLE_GENAI_USE_VERTEXAI=true` and credentials are mounted correctly.

### Issue: Audio not playing

**Check:**
- ElevenLabs API key is valid
- Check browser console for audio playback errors
- Verify CORS settings allow audio blob URLs

### Issue: Progress counter shows wrong numbers

**Fixed in latest version.** If you see "Generating voice audio 5/2", update to latest code.

### Issue: Only female/male voices

**Fixed in latest version.** Voice selection now uses 12 different voices (6 male, 6 female) based on character names.

## Security Considerations

1. **Never commit credentials**:
   - Add `gcp-key.json` to `.gitignore`
   - Use environment variables for all secrets

2. **Use secrets management**:
   - Google Secret Manager for GCP
   - AWS Secrets Manager for AWS
   - HashiCorp Vault for multi-cloud

3. **Enable CORS properly**:
   - Backend should only allow your frontend domain
   - Don't use `*` in production

4. **Rate limiting**:
   - Consider adding rate limits to prevent API abuse
   - Protect expensive endpoints (StartSession, AskQuestion)

## Scaling

The application is stateless and can scale horizontally:

1. **Backend**: Scale based on CPU/memory
   - Each game session generates ~15 cases upfront
   - Most expensive operation: Vertex AI + ElevenLabs calls

2. **Frontend**: Static assets, scales infinitely via CDN

3. **Database**: Currently in-memory
   - For production, migrate to Firestore
   - Enable horizontal scaling

## Next Steps

- [ ] Set up CI/CD pipeline (GitHub Actions example included)
- [ ] Configure monitoring and alerting
- [ ] Implement Firestore for persistent state
- [ ] Add rate limiting and auth
- [ ] Set up CDN for frontend assets
