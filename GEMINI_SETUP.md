# Gemini API Setup Guide

The backend is now fully integrated with Google's Gemini API! This guide will help you get it running.

## What's Been Implemented

All four Gemini service methods are now live:

1. **GenerateRules** - Creates daily transit rules for each game session
2. **GenerateCases** - Generates NPC workers with documents and contradictions
3. **GenerateDialogue** - Powers real-time NPC responses to your questions
4. **GenerateVerdict** - Explains whether your decision was correct/incorrect

The implementation includes:
- Smart fallback to mock data if no API key is provided
- JSON parsing with error handling
- Temperature tuning for creative/consistent outputs
- Proper prompt engineering for each use case

## Prerequisites

✅ **Environment Loading:** The backend now uses `godotenv` to automatically load `.env` files - no manual export needed!

## Getting Your Gemini API Key

### Step 1: Get API Key (Free Tier Available)

1. Go to [Google AI Studio](https://aistudio.google.com/app/apikey)
2. Sign in with your Google account
3. Click **"Get API Key"** or **"Create API Key"**
4. Copy your API key

**Note:** Gemini has a generous free tier that's perfect for development and demos.

### Step 2: Set Up Environment Variable

In the `backend/` directory:

```bash
# Copy the example file
cp .env.example .env

# Edit the .env file and add your API key
# Replace 'your_gemini_api_key_here' with your actual key
```

Your `.env` file should look like:

```
PORT=8080
GEMINI_API_KEY=AIzaSyC-YourActualAPIKeyHere
ELEVENLABS_API_KEY=
FIRESTORE_PROJECT_ID=
```

## Running the Game

### Option 1: With Gemini API (Recommended)

1. Make sure your `.env` file has the API key
2. Start the backend:

```bash
cd backend
go run cmd/server/main.go
```

3. Start the frontend:

```bash
cd frontend
npm run dev
```

4. Open http://localhost:3000

You'll get:
- ✅ Unique rules every game session
- ✅ Varied NPC personalities and cases
- ✅ Dynamic dialogue responses
- ✅ AI-generated verdict explanations

### Option 2: Mock Mode (No API Key)

If you want to test without an API key:

1. Leave `GEMINI_API_KEY` empty in `.env` (or don't create a `.env` file)
2. The backend will automatically use hardcoded mock data
3. Perfect for quick testing or offline development

## How It Works

### Session Start Flow

When you start a new game session:

1. **GenerateRules** is called → Creates 3-5 unique rules
2. **GenerateCases** is called → Generates 15-20 worker cases
3. All data is cached in session state
4. No more API calls until you ask questions

### During Gameplay

- **GetNextCase** → Returns pre-generated case (instant, no API call)
- **AskQuestion** → Calls **GenerateDialogue** in real-time
- **ResolveCase** → Calls **GenerateVerdict** to explain outcome

### API Usage

Approximate API calls per game session:
- 1 call for rules generation
- 1 call for bulk case generation (generates 15-20 cases at once)
- 1 call per question asked (varies by player)
- 1 call per verdict explanation

**Total cost:** Essentially free on Gemini's free tier.

## Prompt Engineering Details

### GenerateRules Prompt
- Requests JSON array of 3-5 rules
- Emphasizes bureaucratic tone
- Focuses on verifiable document fields

### GenerateCases Prompt
- Batch generates multiple cases in one call
- Specifies exact JSON structure
- Requests 60/40 split (approve/deny ratio)
- Includes contradictions for interesting gameplay

### GenerateDialogue Prompt
- Provides full NPC context (personality, demeanor)
- Includes ground truth (so NPC can be evasive about lies)
- Temperature: 1.2 (creative, varied responses)
- Enforces character consistency

### GenerateVerdict Prompt
- Provides case details and player decision
- Requests concise 1-2 sentence explanation
- Temperature: 0.7 (balanced creativity/consistency)
- Professional supervisor tone

## Model Choice

Using **gemini-pro** for all operations:
- Stable and reliable
- Good balance of speed and quality
- Widely available across all API versions
- Great for JSON-structured outputs and creative tasks

## Troubleshooting

### "failed to create Gemini client"

Check that:
- Your API key is in `.env` file
- The key starts with `AIzaSy`
- No extra spaces/quotes around the key

### Getting mock data despite having API key

The system falls back to mocks if:
- API call fails
- Response is empty
- JSON parsing fails

Check backend logs for error messages.

### Rate limiting

Free tier limits:
- 15 requests per minute
- 1,500 requests per day

Our batch case generation helps stay under these limits.

## Testing

### Test with mock data:
```bash
# Don't set GEMINI_API_KEY
cd backend
go run cmd/server/main.go
```

### Test with real API:
```bash
# Set GEMINI_API_KEY in .env
cd backend
go run cmd/server/main.go
```

Watch the logs to see API calls happening.

## Next Steps

Once Gemini is working, you can:
1. Add ElevenLabs for voice generation
2. Implement Firestore for session persistence
3. Tune prompts based on gameplay feedback
4. Experiment with different temperature values

## Files Changed

- [backend/internal/services/gemini/client.go](backend/internal/services/gemini/client.go) - Full Gemini integration
- [backend/go.mod](backend/go.mod) - Added `github.com/google/generative-ai-go`
- [backend/.env.example](backend/.env.example) - Environment template

## Development Notes

The implementation gracefully handles failures:
- No API key? → Uses mocks
- API error? → Fallback to mocks
- Invalid JSON? → Fallback to mocks

This ensures the game always works, even in degraded mode.

---

**Status:** ✅ Complete and tested (compiles, runs in both modes)

**Ready to use:** Just add your API key!
