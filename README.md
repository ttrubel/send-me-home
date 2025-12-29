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

### Backend (.env in backend/)
```
PORT=8080
GEMINI_API_KEY=your_gemini_api_key
ELEVENLABS_API_KEY=your_elevenlabs_api_key
FIRESTORE_PROJECT_ID=your_project_id
```

### Frontend (.env in frontend/)
```
VITE_API_URL=http://localhost:8080
```

## Next Steps

See individual README files for more details:
- [Backend README](./backend/README.md)
- [Frontend README](./frontend/README.md)
- [Game Design Document](./GAME.md)

## TODO

- [ ] Implement real Gemini API integration
- [ ] Implement real Firestore integration
- [ ] Add ElevenLabs voice generation
- [ ] Add document rendering with visuals
- [ ] Add animations and polish
- [ ] Deploy to Google Cloud

## License

MIT
