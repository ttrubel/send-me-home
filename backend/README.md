# Send Me Home - Backend

Go backend service using Connect-RPC for the Send Me Home game.

## Setup

### Prerequisites
- Go 1.21+
- Buf CLI (for code generation)

### Install Dependencies

```bash
cd backend
go mod download
```

### Environment Variables

Create a `.env` file in the backend directory:

```bash
PORT=8080
GEMINI_API_KEY=your_gemini_api_key
ELEVENLABS_API_KEY=your_elevenlabs_api_key
FIRESTORE_PROJECT_ID=your_project_id
```

## Development

### Generate Code from Proto

From the monorepo root:
```bash
buf generate
```

This generates:
- Go server stubs in `backend/gen/`
- TypeScript client in `frontend/src/gen/`

### Run Server

```bash
go run cmd/server/main.go
```

Server runs on `http://localhost:8080`

### API Endpoints

- `POST /game.v1.GameService/StartSession` - Start new game session
- `POST /game.v1.GameService/GetNextCase` - Get next case
- `POST /game.v1.GameService/AskQuestion` - Ask NPC question (streaming)
- `POST /game.v1.GameService/SecondaryCheck` - Perform secondary check
- `POST /game.v1.GameService/ResolveCase` - Submit decision
- `POST /game.v1.GameService/GetSessionStatus` - Get session stats
- `GET /health` - Health check

## Project Structure

```
backend/
├── cmd/
│   └── server/          # Main server entry point
├── internal/
│   ├── api/             # Connect-RPC handlers
│   ├── config/          # Configuration
│   ├── models/          # Data models
│   └── services/
│       ├── gemini/      # Gemini AI client
│       └── firestore/   # Session storage
├── gen/                 # Generated code (do not edit)
└── go.mod
```

## TODO

- [ ] Implement actual Gemini API integration
- [ ] Implement actual Firestore integration
- [ ] Add ElevenLabs voice generation
- [ ] Add proper error handling and validation
- [ ] Add unit tests
- [ ] Add Docker support
