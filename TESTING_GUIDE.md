# Testing the Gemini Integration

Your game is now fully connected! Here's how to test it.

## Quick Start

### 1. Make sure backend is running

```bash
cd backend
go run cmd/server/main.go
```

You should see:
```
Server starting on port 8080
Game service available at http://localhost:8080/game.v1.GameService/
```

### 2. Start the frontend

```bash
cd frontend
npm run dev
```

You should see:
```
VITE ready in XXX ms
Local: http://localhost:3000
```

### 3. Test the game

Open http://localhost:3000 and:

1. Click "START NEW SESSION"
2. Watch the progress bar as it generates cases
3. Play through a few cases

## What to Look For

### With Gemini API Key (Real Mode)

If you added your API key to `backend/.env`:

✅ **Unique rules each session** - Refresh and start a new game, rules will be different
✅ **Varied NPC names** - "Sarah Chen", "Marcus Rodriguez", etc. (not "Worker 1", "Worker 2")
✅ **Dynamic responses** - Ask questions, get contextual answers
✅ **AI verdicts** - Detailed explanations when you approve/deny

**Backend logs will show:**
```
RPC: /game.v1.GameService/StartSession
RPC: /game.v1.GameService/GetNextCase
RPC: /game.v1.GameService/AskQuestion
RPC: /game.v1.GameService/ResolveCase
```

### Without API Key (Mock Mode)

If `GEMINI_API_KEY` is empty in `backend/.env`:

⚠️ **Same rules every session** - Standard 5 rules
⚠️ **Generic NPCs** - "Worker 1", "Worker 2", "Worker 3", etc.
⚠️ **Simple responses** - "I understand your question about '...'."
⚠️ **Basic verdicts** - "Correct! Contract term complete, no incidents..."

This is **expected behavior** - the fallback mocks work perfectly for testing without an API key.

## Testing Checklist

### Basic Flow (Works in Both Modes)
- [ ] Session starts without errors
- [ ] Progress bar shows during generation
- [ ] Rules display in left panel
- [ ] First case loads automatically
- [ ] NPC portrait appears
- [ ] Documents render as paper style
- [ ] Can ask questions (input + Space bar)
- [ ] Can approve/deny workers
- [ ] Verdict modal appears
- [ ] Score updates correctly
- [ ] Queue counter decrements

### Gemini-Specific Features (Only with API Key)
- [ ] Rules are unique each session
- [ ] NPCs have realistic names
- [ ] NPC responses feel natural and in-character
- [ ] Questions about lies get evasive answers
- [ ] Verdict explanations are detailed and contextual

### Performance
- [ ] Session start completes in 5-10 seconds
- [ ] Case loading is instant (pre-generated)
- [ ] Question responses appear within 1-2 seconds
- [ ] No console errors in browser

## Debugging

### Backend Not Starting

```bash
# Check if port 8080 is in use
lsof -i :8080

# Kill any process using it
kill -9 <PID>

# Try again
cd backend && go run cmd/server/main.go
```

### Frontend Can't Connect

Check `frontend/.env`:
```
VITE_API_URL=http://localhost:8080
```

Make sure backend is running first.

### Getting Mock Data Despite Having API Key

Check `backend/.env`:
```bash
cat backend/.env
```

Your `GEMINI_API_KEY` should look like:
```
GEMINI_API_KEY=AIzaSyC-YourActualKeyHere
```

**Common mistakes:**
- Extra quotes around key: ❌ `GEMINI_API_KEY="AIza..."`
- Extra spaces: ❌ `GEMINI_API_KEY= AIza...`
- Wrong key format: ❌ `GEMINI_API_KEY=sk-...` (that's OpenAI)

Restart backend after fixing `.env` file.

### Console Errors

**CORS errors:**
- Backend has CORS enabled for `localhost:3000`, `localhost:3001`, `localhost:5173`
- Make sure frontend is running on one of these ports

**"Session not found":**
- Start a new session from the beginning
- Backend uses in-memory storage (sessions lost on restart)

## API Call Flow

Here's what happens when you play:

### Session Start
```
Frontend → StartSession RPC
  ↓
Backend → gemini.GenerateRules()
  ↓
Backend → gemini.GenerateCases(rules, 15)
  ↓
Backend → Save to Firestore (in-memory)
  ↓
Frontend ← Session ID + Rules + Case count
```

### Playing a Case
```
Frontend → GetNextCase RPC
  ↓
Backend → Load from session state (instant)
  ↓
Frontend ← Case data (NPC, documents, opening line)

Frontend → AskQuestion RPC
  ↓
Backend → gemini.GenerateDialogue(question, npc, truth)
  ↓
Frontend ← Streaming text response

Frontend → ResolveCase RPC
  ↓
Backend → gemini.GenerateVerdict(case, decision)
  ↓
Backend → Update score, increment case index
  ↓
Frontend ← Verdict + Score + Outcome
```

## Network Inspector

Open browser DevTools → Network tab:

You should see requests to:
- `http://localhost:8080/game.v1.GameService/StartSession`
- `http://localhost:8080/game.v1.GameService/GetNextCase`
- `http://localhost:8080/game.v1.GameService/AskQuestion`
- `http://localhost:8080/game.v1.GameService/ResolveCase`

All should return `200 OK`.

## Example Test Session

1. **Start session** - Should take 5-10 seconds
2. **First case loads** - Worker with name, role, documents
3. **Ask question**: "What is your home planet?"
   - With API: Natural response like "Earth, of course. Been mining here for 3 years, can't wait to get back."
   - Mock mode: "I understand your question about 'What is your home planet?'. Let me explain..."
4. **Approve or deny** - Based on rules
5. **Check verdict** - Should explain why correct/incorrect
6. **Next case** - Click "NEXT CASE →"

## Success Indicators

### ✅ Everything Working
- Unique NPCs each session
- Natural dialogue
- Detailed verdicts
- No errors in console
- Smooth gameplay

### ⚠️ Using Mock Data
- Same "Worker 1, Worker 2..." NPCs
- Generic responses
- Simple verdicts
- Still works fine, just not dynamic

### ❌ Something Wrong
- Errors in console
- Connection refused
- Session not found
- CORS errors

## Next Steps

Once confirmed working:
1. Play 2-3 full sessions to test variety
2. Try asking different types of questions
3. Test approve/deny logic with contradictions
4. Check that score calculation is correct
5. Verify all 15 cases work without issues

---

**Current Status:** ✅ Backend integrated, frontend connected, ready to test!

**To switch modes:**
- **Mock mode**: Leave `GEMINI_API_KEY` empty
- **Real mode**: Add your API key to `backend/.env`
