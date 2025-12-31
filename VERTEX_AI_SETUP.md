# Google Gemini Setup Guide

Your backend now uses the **Google Gen AI SDK** (`google.golang.org/genai`) which supports both **Vertex AI** and **AI Studio** with simple environment variable configuration.

## Quick Start (Choose Your Path)

### Option 1: AI Studio (Fastest - 2 minutes)

Perfect for development and demos:

1. Get API key from [Google AI Studio](https://aistudio.google.com/app/apikey)
2. Add to your `backend/.env`:
   ```bash
   GOOGLE_API_KEY=your-api-key-here
   ```
3. Run the backend - that's it!

### Option 2: Vertex AI (Production-ready)

For production deployments with service accounts:

1. Set up Google Cloud project (see detailed steps below)
2. Add to your `backend/.env`:
   ```bash
   GOOGLE_GENAI_USE_VERTEXAI=true
   GOOGLE_CLOUD_PROJECT=your-project-id
   GOOGLE_CLOUD_LOCATION=us-central1
   GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json
   ```
3. Run the backend

### Option 3: Mock Mode (No Setup)

Leave all Gemini environment variables empty - the backend will use hardcoded mock data.

---

## Detailed Setup: AI Studio

### Step 1: Get API Key

1. Go to [Google AI Studio](https://aistudio.google.com/app/apikey)
2. Sign in with your Google account
3. Click **"Get API Key"** or **"Create API Key"**
4. Copy your API key (starts with `AIzaSy...`)

### Step 2: Configure Environment

Create or edit `backend/.env`:

```bash
PORT=8080
GOOGLE_API_KEY=AIzaSyC-YourActualAPIKeyHere
```

### Step 3: Run

```bash
cd backend
go run cmd/server/main.go
```

**That's it!** No Google Cloud project, no service accounts, no OAuth.

### AI Studio Limits

- **Free tier**: 15 requests/minute, 1,500/day
- **Great for**: Development, demos, prototypes
- **Not for**: Production apps with high traffic

---

## Detailed Setup: Vertex AI

### Prerequisites

#### 1. Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Note your **Project ID**

#### 2. Enable Vertex AI API

```bash
gcloud services enable aiplatform.googleapis.com --project=YOUR_PROJECT_ID
```

Or via [Cloud Console](https://console.cloud.google.com/apis/library/aiplatform.googleapis.com)

#### 3. Create Service Account

1. Go to [IAM & Admin > Service Accounts](https://console.cloud.google.com/iam-admin/serviceaccounts)
2. Click **Create Service Account**
3. Name: `send-me-home-gemini`
4. Grant role: **Vertex AI User** (`roles/aiplatform.user`)
5. Click **Done**

#### 4. Download Service Account Key

1. Find your service account in the list
2. Click three dots → **Manage Keys**
3. **Add Key** → **Create New Key**
4. Choose **JSON** format
5. Save the file securely (treat it like a password!)

### Configuration

Create or edit `backend/.env`:

```bash
PORT=8080

# Vertex AI Configuration
GOOGLE_GENAI_USE_VERTEXAI=true
GOOGLE_CLOUD_PROJECT=your-project-id
GOOGLE_CLOUD_LOCATION=us-central1
GOOGLE_APPLICATION_CREDENTIALS=/absolute/path/to/service-account-key.json
```

**Important**:
- Use **absolute paths** for `GOOGLE_APPLICATION_CREDENTIALS`
- Don't commit the JSON key file to git
- Add `*.json` to `.gitignore`

### Running

```bash
cd backend
go run cmd/server/main.go
```

The SDK automatically reads the environment variables and authenticates.

### Vertex AI Locations

Common regions:
- `us-central1` (Iowa, USA)
- `us-east4` (Virginia, USA)
- `us-west1` (Oregon, USA)
- `europe-west4` (Netherlands)
- `asia-southeast1` (Singapore)

[Full list of locations](https://cloud.google.com/vertex-ai/docs/general/locations)

---

## How It Works

The new `google.golang.org/genai` package automatically detects configuration from environment variables:

```go
client, err := genai.NewClient(ctx, &genai.ClientConfig{})
```

This single line works with:
- **AI Studio**: If `GOOGLE_API_KEY` is set
- **Vertex AI**: If `GOOGLE_GENAI_USE_VERTEXAI=true` and project/location are set
- **Mock mode**: If neither is configured

No code changes needed to switch between them!

## Environment Variables Reference

| Variable | Required For | Description |
|----------|--------------|-------------|
| `GOOGLE_API_KEY` | AI Studio | API key from aistudio.google.com |
| `GOOGLE_GENAI_USE_VERTEXAI` | Vertex AI | Set to `true` to use Vertex AI |
| `GOOGLE_CLOUD_PROJECT` | Vertex AI | Your GCP project ID |
| `GOOGLE_CLOUD_LOCATION` | Vertex AI | Region (e.g., `us-central1`) |
| `GOOGLE_APPLICATION_CREDENTIALS` | Vertex AI | Path to service account JSON |
| `GEMINI_MODEL` | Optional | Model to use (default: `gemini-2.5-flash-lite`) |

### Available Models

- **`gemini-2.5-flash-lite`** (default) - Latest Gemini 2.5, optimized for speed and cost
- **`gemini-2.0-flash-exp`** - Experimental Gemini 2.0 (latest features)

Set in your `.env`:
```bash
GEMINI_MODEL=gemini-2.5-flash-lite
```

## Authentication Methods Summary

### AI Studio
- ✅ Simple: Just an API key
- ✅ Fast setup
- ❌ Rate limited
- ❌ Not for production

### Vertex AI
- ✅ Production-ready
- ✅ Higher quotas
- ✅ Enterprise features
- ❌ Requires GCP setup
- ❌ Service account management

### Alternative: gcloud CLI (Development Only)

For local development without service account files:

```bash
gcloud auth application-default login
```

Then just set:
```bash
GOOGLE_GENAI_USE_VERTEXAI=true
GOOGLE_CLOUD_PROJECT=your-project-id
GOOGLE_CLOUD_LOCATION=us-central1
```

The SDK uses your gcloud credentials automatically.

**Note**: This doesn't work in deployed environments - use service accounts for production.

## Troubleshooting

### Error: "API keys are not supported by this API"

**Solution**: You set `GOOGLE_GENAI_USE_VERTEXAI=true` but also have `GOOGLE_API_KEY` set. Either:
- Remove `GOOGLE_GENAI_USE_VERTEXAI` to use AI Studio
- Remove `GOOGLE_API_KEY` and configure Vertex AI properly

### Error: "could not find default credentials"

**Vertex AI**: Make sure `GOOGLE_APPLICATION_CREDENTIALS` points to a valid JSON file with absolute path.

**AI Studio**: Make sure `GOOGLE_API_KEY` is set.

### Error: "Permission denied"

Your service account doesn't have the right role. Add **Vertex AI User**:

```bash
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="serviceAccount:YOUR_SA_EMAIL" \
  --role="roles/aiplatform.user"
```

### Getting mock data despite valid credentials

Check backend logs for error messages. The system falls back to mocks if:
- Client creation fails
- API calls fail
- Response parsing fails

## Cost Estimates

### AI Studio (Free Tier)
- 15 requests/minute
- 1,500 requests/day
- **Cost**: Free

### Vertex AI Pricing (Gemini 1.5 Flash)

- **Input**: $0.000075 per 1K characters
- **Output**: $0.0003 per 1K characters

**Per game session** (~15 cases):
- Rules generation: $0.0003
- Case generation: $0.003
- Q&A (5 questions): $0.0006
- Verdicts: $0.0009
- **Total**: ~$0.005 (half a cent)

**Monthly (1000 games)**: ~$5

Vertex AI also has free tier quotas for testing.

## Security Best Practices

### ❌ Don't
- Commit `.env` files to git
- Commit `*.json` service account keys
- Share API keys publicly
- Use production credentials in development
- Hardcode credentials in code

### ✅ Do
- Add `.env` and `*.json` to `.gitignore`
- Use separate service accounts for dev/prod
- Rotate keys regularly
- Use minimal permissions (Vertex AI User only)
- Store credentials in secrets manager for production

### Recommended `.gitignore` entries

```
backend/.env
backend/*.json
*-service-account-*.json
service-account-*.json
```

## Files Changed

- [backend/internal/services/gemini/client.go](backend/internal/services/gemini/client.go) - Updated to use `google.golang.org/genai`
- [backend/internal/config/config.go](backend/internal/config/config.go) - Removed Gemini config fields (now env vars only)
- [backend/cmd/server/main.go](backend/cmd/server/main.go) - Updated client initialization
- [backend/.env.example](backend/.env.example) - Updated with new env vars
- [backend/go.mod](backend/go.mod) - Added `google.golang.org/genai` dependency

## Testing Your Setup

### Test with mock data:
```bash
# Don't set any Gemini env vars
cd backend
go run cmd/server/main.go
```

### Test with AI Studio:
```bash
# Set GOOGLE_API_KEY in .env
cd backend
go run cmd/server/main.go
```

### Test with Vertex AI:
```bash
# Set GOOGLE_GENAI_USE_VERTEXAI and related vars in .env
cd backend
go run cmd/server/main.go
```

Watch the logs to see which mode it's using.

## What's Next?

Once Gemini is working:
1. ✅ You're done with Gemini setup!
2. Add ElevenLabs for voice (optional)
3. Add Firestore for persistence (optional)
4. Deploy to production

---

**Status**: ✅ Complete and ready to use

**Package**: `google.golang.org/genai` v1.40.0

**Supports**: AI Studio + Vertex AI + Mock mode

Need help? Check the [Google Gen AI SDK docs](https://pkg.go.dev/google.golang.org/genai)
