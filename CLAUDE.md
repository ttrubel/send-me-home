# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Send Me Home** is a voice-driven document inspection game (Papers, Please-style) set at an asteroid mining station. Players verify workers' documents to approve or deny shuttle boarding. The project is a monorepo with Go backend and React TypeScript frontend, communicating via Connect-RPC (gRPC over HTTP/2).

## Core Architecture

### Communication Pattern: Connect-RPC

The entire API is defined in a single proto file (`proto/game/v1/game.proto`). All communication between frontend and backend flows through auto-generated clients. **Never manually write API client code** - always modify the proto and regenerate.

After any proto changes, run:
```bash
buf generate
```

This generates:
- Go server stubs in `backend/gen/`
- TypeScript client in `frontend/src/gen/`

**Important:** The Go package path in proto must match your actual module path. Check `backend/go.mod` for the module name and update `option go_package` in the proto if needed.

### Game Flow Architecture: Pre-generation Strategy

The game follows a hybrid approach optimized for demo performance:

**At Session Start (StartSession RPC):**
- Generate ALL cases upfront (15-20 NPCs with documents)
- Generate opening line audio for each NPC (when ElevenLabs integrated)
- Store everything in session state
- Stream progress updates to client

**During Gameplay:**
- GetNextCase is instant (pulls from pre-generated queue)
- AskQuestion generates real-time responses (can't predict questions)
- ResolveCase validates decision and advances to next case

This design ensures smooth case transitions while keeping Q&A dynamic.

### Data Flow

1. **Session State:** All game state lives in `backend/internal/services/firestore/client.go` (currently in-memory map, designed for easy Firestore swap)
2. **Case Data:** Defined in `backend/internal/models/case.go` - includes NPC profile, documents, truth, and contradictions
3. **Handler Layer:** `backend/internal/api/handler.go` implements all 6 RPC methods and orchestrates services

### Service Layer Pattern

Backend services are clean interfaces designed for easy swapping:

- **gemini.Client** - Currently returns mock data, ready for Gemini API
- **firestore.Client** - In-memory map with Firestore-compatible interface
- **elevenlabs.Client** - Not yet created (TODO)

When implementing real services, maintain the same interface - handlers shouldn't change.

## Development Commands

### Initial Setup
```bash
./setup.sh  # Installs deps, generates code
```

### Code Generation (after proto changes)
```bash
buf generate  # Must run from project root
```

### Running Locally

**Terminal 1 - Backend:**
```bash
cd backend
go run cmd/server/main.go
# Runs on http://localhost:8080
```

**Terminal 2 - Frontend:**
```bash
cd frontend
npm run dev
# Runs on http://localhost:3000
```

Or use Make:
```bash
make dev-backend   # Terminal 1
make dev-frontend  # Terminal 2
```

### Testing

**Backend:**
```bash
cd backend
go test ./...                           # All tests
go test ./internal/services/gemini      # Specific package
go test -v -run TestGenerateCases       # Specific test
```

**Frontend:**
```bash
cd frontend
npm test                     # Watch mode
npm test -- --watchAll=false # Run once
```

### Building

```bash
make build
# Or separately:
cd backend && go build -o bin/server cmd/server/main.go
cd frontend && npm run build
```

## Key Implementation Details

### Streaming RPCs

Two methods use streaming:

**1. StartSession** - Server streams progress updates:
```go
stream.Send(&gamev1.StartSessionResponse{
    Update: &gamev1.StartSessionResponse_Progress{...}
})
// Finally send ready signal
stream.Send(&gamev1.StartSessionResponse{
    Update: &gamev1.StartSessionResponse_Ready{...}
})
```

**2. AskQuestion** - Server streams text then audio (when ElevenLabs integrated):
```go
stream.Send(&gamev1.AskQuestionResponse{
    Chunk: &gamev1.AskQuestionResponse_TextChunk{...}
})
// TODO: Stream audio chunks
stream.Send(&gamev1.AskQuestionResponse{
    Chunk: &gamev1.AskQuestionResponse_Done{Done: true}
})
```

Frontend consumes with async iteration:
```typescript
for await (const response of gameClient.startSession({...})) {
    if (response.update.case === 'progress') { /* ... */ }
    if (response.update.case === 'ready') { /* ... */ }
}
```

### Session State Management

Sessions are identified by UUID. All RPCs (except StartSession) require `session_id`.

**State mutations:**
- `StartSession` creates session
- `GetNextCase` reads current case (no mutation)
- `AskQuestion` reads case data (no mutation)
- `SecondaryCheck` decrements quota
- `ResolveCase` updates score, increments case index

**Important:** `ResolveCase` advances `CurrentCaseIndex` - the next `GetNextCase` returns the following case.

### Scoring System

From `ResolveCase` in `backend/internal/api/handler.go`:
- Correct decision: +10 points
- Incorrect decision: -15 points
- Secondary check: costs 1 quota (no point penalty)

Initial quota: 3 secondary checks per session.

## Integration Points (TODOs)

### Gemini AI Integration

File: `backend/internal/services/gemini/client.go`

Replace mock implementations:
- `GenerateRules()` - Generate 3-5 daily rules
- `GenerateCases()` - Generate NPCs with documents and contradictions
- `GenerateDialogue()` - Respond to player questions in character
- `GenerateVerdict()` - Explain correct/incorrect decision

Use Vertex AI SDK or REST API. Consider parallel generation with goroutines for `GenerateCases()`.

### ElevenLabs Voice Integration

Create: `backend/internal/services/elevenlabs/client.go`

Needed methods:
- `TextToSpeech(voiceID, text string) ([]byte, error)` - For opening lines
- `TextToSpeechStream(voiceID, text string) (io.ReadCloser, error)` - For Q&A

Update `StartSession` to generate opening audio for all cases.
Update `AskQuestion` to stream audio chunks after text.

Frontend needs audio playback in `GameDesk.tsx` - handle audio chunks from streaming response.

### Firestore Integration

File: `backend/internal/services/firestore/client.go`

Replace in-memory map with Firestore SDK. The interface is already compatible - initialize `firestore.Client` in `NewClient()` and implement the same methods using actual Firestore operations.

## Frontend Architecture

### Component Structure

- **App.tsx** - Game state machine (start → playing → complete)
- **SessionStart.tsx** - Handles session initialization with progress bar
- **GameDesk.tsx** - Main gameplay UI (document inspection, Q&A, decisions)

### API Client Pattern

File: `frontend/src/api/client.ts`

Single typed client instance:
```typescript
import { gameClient } from '../api/client';

// All RPC methods available with full type safety
const response = await gameClient.getNextCase({ sessionId });
```

Never import from `gen/` directly - always use the `gameClient` instance.

### State Management

No external state library. State flows:
- App.tsx holds session data
- GameDesk.tsx holds current case and UI state
- API responses drive state transitions

## Common Patterns

### Adding a New RPC Method

1. Add to `proto/game/v1/game.proto`
2. Run `buf generate`
3. Implement in `backend/internal/api/handler.go`
4. Backend signature for unary: `func (h *GameHandler) MethodName(ctx context.Context, req *connect.Request[gamev1.RequestType]) (*connect.Response[gamev1.ResponseType], error)`
5. Backend signature for streaming: `func (h *GameHandler) MethodName(ctx context.Context, req *connect.Request[gamev1.RequestType], stream *connect.ServerStream[gamev1.ResponseType]) error`
6. Frontend auto-gets typed method on `gameClient`

### Adding a New Document Type

1. Update mock generation in `backend/internal/services/gemini/client.go`
2. Add fields to the Document in the mock
3. Frontend automatically renders via `doc.fields` map
4. When Gemini integrated, update case generation prompt

### Modifying Module Path

If you need to change the Go module path:
1. Update `backend/go.mod` module name
2. Update `option go_package` in `proto/game/v1/game.proto`
3. Run `buf generate`
4. Update all imports in Go files

## Environment Variables

Backend (`.env` in `backend/`):
```
PORT=8080
GEMINI_API_KEY=your_key
ELEVENLABS_API_KEY=your_key
FIRESTORE_PROJECT_ID=your_project
```

Frontend (`.env` in `frontend/`):
```
VITE_API_URL=http://localhost:8080
```

## Current Status

**Functional:** Demo works end-to-end with mock data (15 cases, Q&A, scoring, verdict)

**Not Yet Implemented:**
- Real Gemini API calls (using mocks)
- Real Firestore (using in-memory)
- ElevenLabs voice generation
- Audio playback on frontend
- Secondary check UI modal
- Document drag-and-drop
- Animations/polish

See `IMPLEMENTATION.md` for detailed implementation status and `GAME.md` for game design specification.
