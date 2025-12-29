# Quick Start Guide

Get the Send Me Home game running in 5 minutes.

## Step 1: Install Prerequisites

### Required:
- **Go 1.21+**: Download from [go.dev](https://go.dev/dl/)
- **Node.js 18+**: Download from [nodejs.org](https://nodejs.org/)
- **Buf CLI**: Install with `go install github.com/bufbuild/buf/cmd/buf@latest`

### Verify Installation:
```bash
go version    # Should show 1.21 or higher
node --version # Should show 18 or higher
buf --version  # Should show buf version
```

## Step 2: Run Setup Script

From the project root:

```bash
./setup.sh
```

This automatically:
- Installs Go dependencies
- Installs npm dependencies
- Generates code from proto files

## Step 3: Start the Backend

Open a terminal and run:

```bash
cd backend
go run cmd/server/main.go
```

You should see:
```
Server starting on port 8080
Game service available at http://localhost:8080/game.v1.GameService
```

## Step 4: Start the Frontend

Open another terminal and run:

```bash
cd frontend
npm run dev
```

You should see:
```
  VITE ready in 500 ms
  ➜  Local:   http://localhost:3000/
```

## Step 5: Play the Game

1. Open [http://localhost:3000](http://localhost:3000) in your browser
2. Click "Start Shift" to generate game cases
3. Wait for case generation (10-15 seconds)
4. Inspect documents and make decisions!

## Troubleshooting

### "buf: command not found"

Install Buf CLI:
```bash
go install github.com/bufbuild/buf/cmd/buf@latest
```

Make sure `$GOPATH/bin` is in your PATH.

### Port 8080 or 3000 already in use

Change the ports in:
- Backend: Set `PORT` env variable or edit [backend/internal/config/config.go](backend/internal/config/config.go)
- Frontend: Edit `server.port` in [frontend/vite.config.ts](frontend/vite.config.ts)

### "Cannot find module './gen/game/v1/game_pb'"

Run code generation:
```bash
buf generate
```

### Backend crashes on startup

Make sure you're in the backend directory:
```bash
cd backend
go run cmd/server/main.go
```

## Development Workflow

### Making Changes to the API

1. Edit [proto/game/v1/game.proto](proto/game/v1/game.proto)
2. Run `buf generate` from project root
3. Restart backend and frontend

### Making Changes to Backend

1. Edit Go files in `backend/`
2. Restart backend (Ctrl+C and re-run)
3. Changes take effect immediately

### Making Changes to Frontend

1. Edit TypeScript/React files in `frontend/src/`
2. Changes hot-reload automatically (no restart needed)

## Next Steps

- Read [GAME.md](GAME.md) for game design details
- Check [backend/README.md](backend/README.md) for backend architecture
- Check [frontend/README.md](frontend/README.md) for frontend details
- Add Gemini API key to implement real AI case generation
- Add ElevenLabs API key to enable voice synthesis

## API Keys Setup (Optional)

Create `backend/.env`:
```bash
GEMINI_API_KEY=your_gemini_key_here
ELEVENLABS_API_KEY=your_elevenlabs_key_here
```

Restart backend after adding keys.
