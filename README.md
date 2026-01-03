# Send Me Home

A voice-driven space transit desk game powered by AI.

## Prerequisites

- **Go 1.25+**: Download from [go.dev](https://go.dev/dl/)
- **Node.js 18+**: Download from [nodejs.org](https://nodejs.org/)
- **Buf CLI**: Install with `go install github.com/bufbuild/buf/cmd/buf@latest`

## Quick Start

### 1. Automated Setup

Run the setup script from the project root:

```bash
./setup.sh
```

This will:
1. Install backend and frontend dependencies.
2. Generate code from the proto definitions.

### 2. Start the Servers

#### Using `make` commands (recommended)

In separate terminals, run:
```bash
make dev-backend
```
```bash
make dev-frontend
```

The backend will run on `http://localhost:8080` and the frontend on `http://localhost:3000`.

### 3. Play the Game

1. Open [http://localhost:3000](http://localhost:3000) in your browser.
2. Click "Start Shift" to generate game cases.
3. Inspect documents and make your decisions!

## Architecture

- **Backend:** A Go service using [Connect-RPC](https://connectrpc.com/) for type-safe, high-performance APIs. It integrates with Gemini for AI-driven content and ElevenLabs for voice synthesis.
- **Frontend:** A React application built with TypeScript and Vite. It uses a generated Connect-RPC client to communicate with the backend.
- **API:** The entire API is defined using Protocol Buffers in `proto/game/v1/game.proto`. This single source of truth is used to generate both the Go server implementation and the TypeScript client.

## API Endpoints

The game uses the following RPCs:

- `StartSession`: Initializes a new game session, generating all cases upfront.
- `GetNextCase`: Fetches the next case for the player to review.
- `AskQuestion`: Allows the player to ask questions to the NPC, returning a streaming response with text and audio.
- `SecondaryCheck`: Performs a secondary verification check on a document.
- `ResolveCase`: Submits the player's decision (approve or deny) for a case.
- `GetSessionStatus`: Retrieves the current session status and score.

## Environment Variables

Create a `.env` file in the project root by copying `.env.example`.

```bash
# Google Cloud / Vertex AI Configuration
# Set GOOGLE_GENAI_USE_VERTEXAI to "false" to use mock data instead of the live API
GOOGLE_GENAI_USE_VERTEXAI=true
GOOGLE_CLOUD_PROJECT=your-gcp-project-id
GCP_CREDENTIALS_PATH=./gcp-key.json

# ElevenLabs API Key for voice generation
ELEVENLABS_API_KEY=your-elevenlabs-api-key

# URL for the frontend to connect to the backend API
VITE_API_URL=http://localhost:8080
```

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
- Secondary check UI modal
- Document drag-and-drop
- Animations/polish

## Performance

- **Cost per game:** ~$0.06-0.12 (Gemini + ElevenLabs)
- **Session generation:** ~30-45 seconds for 15 cases
- **Response time:** <2s for dialogue, instant for case loading

## License

MIT
