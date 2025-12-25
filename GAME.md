# Send Me Home — Voice-Driven Space Transit Desk Game (Hackathon Spec)

## 1) One-line pitch
You’re the **Transit Clerk** at a remote asteroid mine. A line of exhausted workers tries to board the **last shuttle home**. You must **Approve / Deny / Secondary Check** by inspecting **contracts, shift logs, and clearance badges**—while they **talk back** with fully voiced, AI-generated personalities.

---

## 2) Why this idea works (hackathon + judging criteria)
### Technological Implementation
- **Gemini / Vertex AI** generates each NPC’s structured “case file” (public story + hidden truth + planned inconsistencies), drives dialog responses, and produces an explainable verdict.
- **ElevenLabs** provides character voice acting (distinct voices/personas) and the “show” feel.
- **Deterministic rendering** of documents from JSON (clean, legible, consistent).
- **Stateful orchestration** (session/day rules, queue pressure, scoring, quotas) on Google Cloud (Cloud Run + Firestore).

### Design
- Tactile “desk” UI: draggable docs, stamp buttons, rulebook cards, timer/queue meter.
- Clear loop: read rules → inspect docs → ask questions → decide → feedback.

### Potential Impact
- Highly replayable entertainment with low friction (voice makes it feel alive).
- Strong demo value: procedural narrative + game mechanics (not “just a chatbot”).

### Quality of the Idea
- Distinct twist: **blue-collar space dystopia comedy** vs typical “border/customs” clones.
- AI is native to the premise: endless unique NPCs and contradictions.

---

## 3) Setting and tone
- **Setting:** Company-controlled **Transit Gate** at an asteroid mine.
- **Tone:** Corporate bureaucracy comedy + human drama (tired workers, shady supervisors, absurd rules).
- **Safety framing:** Fictional corp IDs and permits (avoid real-world passport vibes).

---

## 4) Core gameplay loop
1. **Rule of the Day** appears (3–5 rules).
2. NPC arrives and speaks (voiced) + presents documents.
3. Player can:
   - Inspect and compare docs
   - Ask 1–3 questions
   - Use **Secondary Check** (limited)
4. Player chooses: ✅ Approve / ❌ Deny / 🟨 Secondary Check
5. Result screen: correctness, reason, and clear explanation of mismatches.
6. End of “shift”: performance summary.

---

## 5) Minimal document set (MVP-friendly)
Use **3 docs** (enough contradictions without complexity):

1) **Employment Contract**
- Name / Employee ID
- Role + Department
- Term end date
- Signatures / corp seal (fictional)

2) **Shift Log / Hours Sheet**
- Last shift date/time
- Total hours
- Incident flags / missed debrief

3) **Clearance Badge**
- Access level (A/B/C)
- Medical clearance status
- Zone authorization

**Optional later:** Shuttle ticket / seat assignment.

---

## 6) Contradiction patterns (reusable)
- Contract term ends later, but NPC claims “term complete today.”
- Shift log shows missed debrief; badge says “cleared.”
- Department/role mismatch between contract and badge level.
- Name or employee ID differs across docs (subtle typo).
- Incident flag present but NPC insists “clean record.”
- Medical clearance required by rule, but badge is expired.

---

## 7) Secondary Check (what it is + how to implement)
### Meaning
A **third decision** that buys **extra verification** when you’re unsure—at a cost. It prevents coin-flip decisions and adds strategy.

### Implementation (simple MVP)
- Player has a **daily quota**: e.g., **3 Secondary Checks per shift**.
- Clicking **Secondary** opens a small panel that runs **one verification tool** instantly, then forces a final Approve/Deny.

### Recommended single tool for MVP
**Contract Verification Terminal**
- Input: Employee ID / Contract ID
- Output: `valid/invalid` + one decisive field (e.g., official term end date)

### Cost options (pick one)
- **Quota cost (recommended):** 3 per day.
- **Time cost:** adds +15 seconds to clock.
- **Score cost:** small penalty per use.

### Important stability rule
Generate `secondary_results` **at case creation time** (deterministic), so secondary checks are instant and consistent (no “LLM changed its mind”).

---

## 8) Scoring (fair + gamey)
### Per NPC (example)
- +10 correct decision
- -15 incorrect decision
- -1 per extra second beyond target (or a “Queue Pressure” meter)
- Secondary check: consumes quota (or adds time)

### Optional bonus (nice UI)
+5 if the player taps/cites the exact discrepancy (“Term end date mismatch”).

---

## 9) “Rule of the Day” system (progression engine)
Each shift begins with 3–5 rules, e.g.:
- “Only workers with **term complete** may board.”
- “Any **incident flag** requires supervisor sign-off.”
- “Medical clearance required after **exposure**.”
- “No unpaid equipment fees.”

Gemini can generate daily rules from an allowed template set, and cases are generated to comply or violate specific rules.

---

## 10) AI roles (clean separation)
### A) Director / Case Generator (Gemini on Vertex AI)
Outputs structured JSON:
- `public_profile`: name, role, demeanor, quirks
- `docs`: contract, shift log, badge (display fields)
- `truth`: hidden ground truth
- `contradictions`: list of intended mismatches (for explainability)
- `secondary_results`: deterministic results for each tool
- `dialogue_style`: talkative/evasive/angry/funny

### B) NPC Dialogue (Gemini)
Given player question + case truth, returns short in-character answers that:
- Maintain consistency with `truth`
- Lie only according to `lie_plan`
- Avoid drifting into random new facts

### C) Verdict (Gemini)
Explains outcome:
- “You approved but term ends next week (contract vs terminal check).”
- Quotes specific fields/turns.

### D) Voice (ElevenLabs)
- Host line: “Next!”
- NPC line delivery (distinct persona)
- Optional supervisor voice in Secondary (later)

---

## 11) Technical architecture (hackathon-realistic)
### Frontend (React)
- Desk UI: draggable documents, rulebook, stamp buttons.
- NPC panel: portrait + subtle motion.
- Audio: plays voiced NPC lines; captions for accessibility.

### Backend (Google Cloud Run)
- `POST /session/start` → creates shift rules + session state (Firestore)
- `GET /case/next` → returns next NPC case JSON + portrait URL + first spoken line
- `POST /case/ask` → player question → Gemini reply text → ElevenLabs voice
- `POST /case/secondary` → returns precomputed secondary result
- `POST /case/resolve` → approve/deny → scoring + verdict
- Persist everything in Firestore: session, cases, decisions.

### Data storage
- Firestore: session state, case files, transcript, outcomes.
- Cloud Storage: optional cached portraits.

### Reliability must-haves
- Strict JSON schema + validation (reject/repair invalid JSON).
- Deterministic doc rendering from JSON (no image-text generation).
- Prefetch next case while current NPC speaks (reduces perceived latency).

---

## 12) Visual design (2D recommended)
### Why 2D
Fast to build, perfect for “document inspection,” easy to polish.

### Core screen layout
- **Top:** Rule of the Day + timer/queue meter
- **Center:** Desk with 3 docs (contract, shift, badge)
- **Right:** Approve / Deny / Secondary buttons + Secondary quota
- **Left:** NPC portrait panel + subtitle transcript

### Polish (cheap but high impact)
- Stamp animation + paper “thunk”
- Scanner light (green/red)
- Field highlight when comparing docs (click a field → matching fields glow)

---

## 13) Character animation approach (simple)
Skip lip-sync. Use subtle motion:
- **Parallax NPC panel** (layered image offsets via pointer/idle drift)
- Optional: tiny idle bob + lighting gradient shift

---

## 14) MVP plan (what to build first)
### MVP v1 (demo-ready)
- One shift with 10–20 procedurally generated NPCs
- 3 documents
- 3–5 daily rules
- Approve/Deny + end-of-case verdict
- ElevenLabs voiced NPC
- Deterministic doc rendering from JSON
- Firestore session state

### MVP v2 (if time)
- Secondary Check (quota: 3/day) with Contract Verification Terminal
- Mismatch highlighting UI
- Supervisor voice line in secondary (optional)

### Stretch goals
- More doc types (ticket, equipment fee bill)
- Difficulty scaling (fewer secondary checks, stricter rules)
- Consequence system (reputation, fines, “stranded workers” narrative)

---

## 15) Project description (submission-ready)
**Send Me Home** is a voice-driven, procedural “document-checking” game set at a remote asteroid mine’s last shuttle gate. Players act as the transit clerk and must approve, deny, or send workers to secondary checks by inspecting contracts, shift logs, and clearance badges under changing “Rule of the Day” policies. **Gemini on Vertex AI** generates each NPC’s structured case file (public story + hidden truth), drives in-character interrogation responses, and delivers transparent verdicts that cite the exact inconsistencies. **ElevenLabs** gives each character a distinct, expressive voice and personality, turning every encounter into a comedic mini-drama. The experience is designed as a tactile desk UI with draggable documents, mismatch highlighting, queue pressure, and satisfying stamp/scanner feedback. The backend runs on **Google Cloud Run** with persistent session state in **Firestore**, showcasing robust multi-service orchestration, validation, and low-latency interaction in a highly replayable game loop.
