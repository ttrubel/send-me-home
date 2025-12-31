# Send Me Home

A voice-driven space transit desk game powered by AI.

## Project Structure

This is a monorepo containing:

- **backend/** - Go backend service (Google Cloud Run)
- **frontend/** - React TypeScript frontend
- **proto/** - Protocol Buffer definitions
- **docs/** - Additional documentation

## Prerequisites

- Go 1.21+
- Node.js 18+ and npm
- Buf CLI (for proto code generation)
- Make (optional, for convenience commands)

## Quick Start

### Automated Setup

```bash
./setup.sh
```

This will:
1. Install backend dependencies
2. Install frontend dependencies
3. Generate code from proto definitions

### Manual Setup

1. **Generate code from proto:**
```bash
buf generate
```

2. **Start backend:**
```bash
cd backend
go run cmd/server/main.go
```

3. **Start frontend (in another terminal):**
```bash
cd frontend
npm run dev
```

## Deployment

### Quick Deployment

Use the interactive deployment script:

```bash
./deploy.sh
```

This will guide you through:
1. Local deployment with Docker Compose
2. Deployment to Google Cloud Run
3. Building Docker images only

### Docker Compose (Local)

```bash
# Configure environment
cp .env.example .env
# Edit .env with your credentials

# Start application
make docker-up

# View logs
docker-compose logs -f

# Stop
make docker-down
```

Access: http://localhost:8080

**Note:** Single unified container serves both frontend and API.

### Google Cloud Run (Production)

```bash
# Set your project ID
export GCP_PROJECT_ID=your-project-id

# Deploy
make deploy-gcp
```

See [DEPLOY_SIMPLE.md](./DEPLOY_SIMPLE.md) for simplified deployment guide or [DEPLOYMENT.md](./DEPLOYMENT.md) for advanced options.

## Development

### Using Make Commands

```bash
# Install all dependencies
make install

# Run backend (terminal 1)
make dev-backend

# Run frontend (terminal 2)
make dev-frontend

# Build everything
make build

# Run tests
make test

# Clean build artifacts
make clean

# Docker commands
make docker-build    # Build images
make docker-up       # Start services
make docker-down     # Stop services
```

### API Documentation

The game uses Connect-RPC (gRPC-compatible) over HTTP/2.

**Available endpoints:**
- `POST /game.v1.GameService/StartSession` - Start new game session (streaming)
- `POST /game.v1.GameService/GetNextCase` - Get next case
- `POST /game.v1.GameService/AskQuestion` - Ask NPC question (streaming)
- `POST /game.v1.GameService/SecondaryCheck` - Perform verification check
- `POST /game.v1.GameService/ResolveCase` - Submit decision
- `POST /game.v1.GameService/GetSessionStatus` - Get session stats

### Project Structure

```
send-me-home/
├── backend/              # Go backend
│   ├── cmd/server/       # Server entry point
│   ├── internal/
│   │   ├── api/          # Connect-RPC handlers
│   │   ├── models/       # Data models
│   │   ├── services/     # Gemini, Firestore, ElevenLabs
│   │   └── config/       # Configuration
│   └── gen/              # Generated Go code
├── frontend/             # React TypeScript frontend
│   ├── src/
│   │   ├── components/   # React components
│   │   ├── api/          # API client
│   │   └── gen/          # Generated TS code
│   └── public/
├── proto/                # Protocol Buffer definitions
│   └── game/v1/
│       └── game.proto
├── buf.yaml              # Buf configuration
├── buf.gen.yaml          # Code generation config
├── Makefile              # Build commands
└── setup.sh              # Setup script
```

## Architecture

- **Backend:** Go service with Connect-RPC
  - Gemini AI for case generation and dialogue
  - ElevenLabs for voice synthesis
  - Firestore for session state (in-memory for now)

- **Frontend:** React + TypeScript + Vite
  - Connect-RPC client (auto-generated)
  - Real-time streaming support
  - Document inspection UI

- **API:** Connect-RPC over HTTP/2
  - Type-safe communication
  - Streaming support for audio/progress
  - Single proto definition

## Environment Variables

Create a `.env` file in the project root (see `.env.example`):

```bash
# Google Cloud / Vertex AI Configuration
GOOGLE_GENAI_USE_VERTEXAI=true
GOOGLE_CLOUD_PROJECT=your-gcp-project-id
GCP_CREDENTIALS_PATH=./gcp-key.json

# ElevenLabs API Key
ELEVENLABS_API_KEY=your-elevenlabs-api-key

# Firestore (optional)
FIRESTORE_PROJECT_ID=your-firestore-project-id

# Frontend API URL
VITE_API_URL=http://localhost:8080
```

For development without API keys, set `GOOGLE_GENAI_USE_VERTEXAI=false` to use mock data.

## Documentation

- [Deployment Guide](./DEPLOYMENT.md) - Detailed deployment instructions
- [Game Design Document](./GAME.md) - Game mechanics and design
- [Implementation Details](./IMPLEMENTATION.md) - Technical implementation
- [Vertex AI Setup](./VERTEX_AI_SETUP.md) - Gemini AI integration
- [Audio Implementation](./AUDIO_IMPLEMENTATION.md) - Voice and audio system

## Features

✅ **Implemented:**
- AI-generated cases with Gemini (Vertex AI)
- Voice-acted NPCs with ElevenLabs (12 different voices)
- Emotional voice delivery based on outcomes
- Dynamic NPC reactions (thank you messages / angry insults)
- Real-time dialogue with streaming responses
- Document inspection gameplay
- Scoring and accuracy tracking
- 15 cases per game session

🔄 **In Progress:**
- Firestore integration for persistent sessions
- Advanced document visuals
- Additional game mechanics

## Performance

- **Cost per game:** ~$0.06-0.12 (Gemini + ElevenLabs)
- **Session generation:** ~30-45 seconds for 15 cases
- **Response time:** <2s for dialogue, instant for case loading

## License

MIT
