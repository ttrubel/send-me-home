package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/ttrubel/send-me-home/internal/models"
	"google.golang.org/genai"
)

type Client struct {
	client *genai.Client
	model  string
}

func NewClient() *Client {
	// Get model from environment variable, default to gemini-2.5-flash-lite
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.5-flash-lite"
	}

	return &Client{
		model: model,
	}
}

// ptr is a helper to create a pointer to a value
func ptr[T any](v T) *T {
	return &v
}

// initClient initializes the Gemini client if not already initialized
func (c *Client) initClient(ctx context.Context) error {
	if c.client != nil {
		return nil
	}

	// Use environment variables for configuration:
	// - For Vertex AI: Set GOOGLE_GENAI_USE_VERTEXAI=true, GOOGLE_CLOUD_PROJECT, GOOGLE_CLOUD_LOCATION
	// - For AI Studio: Set GOOGLE_API_KEY
	// If neither is set, use mock mode
	client, err := genai.NewClient(ctx, &genai.ClientConfig{})
	if err != nil {
		// If client creation fails, use mock mode
		return nil
	}

	c.client = client
	return nil
}

// GenerateRules generates daily rules for the shift
func (c *Client) GenerateRules(ctx context.Context, gameDate string) ([]string, error) {
	if err := c.initClient(ctx); err != nil {
		return nil, err
	}

	// Fallback to mock if no client
	if c.client == nil {
		return c.mockRules(), nil
	}

	prompt := fmt.Sprintf(`You are generating rules for a Papers, Please-style game set on an asteroid mining station.

TODAY'S GAME DATE: %s

Generate 3-4 daily transit rules that workers must comply with to board the final departure shuttle.

Workers have TWO documents:
1. EMPLOYEE BADGE: name, picture URL, job title, issue date, expire date, company name
2. CLEARANCE FORM: name, shift_status (one of: "COMPLETE", "INCOMPLETE", "OVERTIME"), cargo items (cargo1, cargo2)

SHIFT STATUS (be clear):
- "COMPLETE" = Worker finished their shift, can go home
- "INCOMPLETE" = Worker didn't finish shift, should be denied
- "OVERTIME" = Worker did extra hours (can still go home if rules allow)

CARGO CATEGORIES (be very specific):
- ALLOWED: "Personal clothing", "Family photos", "Toiletries", "Snacks", "Music player", "Books", "Personal tablet"
- COMPANY PROPERTY (forbidden): "Delta-7 drill bit", "Company tablet", "Mining helmet", "Safety equipment", "Company radio", "Work tools"
- CONTRABAND (forbidden): "Asteroid samples", "Ore samples", "Minerals", "Live specimens", "Alcohol", "Weapons"

Rules should:
- Be SHORT (one sentence max, under 60 characters ideal)
- Be VERY SPECIFIC about what's allowed/forbidden
- Create clear violations (no ambiguity)
- Relate to: badge expiration (check against today's date %s), cargo restrictions, shift completion status

Examples of good rules:
- "Only COMPLETE shifts can board"
- "No company tools leave the station"
- "Expired badges = denied, no exceptions"
- "Personal items only - no ore samples"
- "INCOMPLETE shifts stay on station"

Return ONLY a JSON array of strings, no other text:
["rule 1", "rule 2", "rule 3", "rule 4"]`, gameDate, gameDate)

	genConfig := &genai.GenerateContentConfig{
		Temperature: ptr(float32(1.0)),
	}

	resp, err := c.client.Models.GenerateContent(ctx, c.model, genai.Text(prompt), genConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate rules: %w", err)
	}

	text := resp.Text()
	if text == "" {
		return c.mockRules(), nil
	}

	// Extract JSON from response
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```json") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
	}

	var rules []string
	if err := json.Unmarshal([]byte(text), &rules); err != nil {
		return c.mockRules(), nil
	}

	return rules, nil
}

// GenerateCases generates multiple cases in parallel
func (c *Client) GenerateCases(ctx context.Context, rules []string, count int, gameDate string) ([]models.Case, error) {
	if err := c.initClient(ctx); err != nil {
		return nil, err
	}

	// Fallback to mock if no client
	if c.client == nil {
		return c.mockCases(count, gameDate), nil
	}

	rulesText := strings.Join(rules, "\n- ")

	prompt := fmt.Sprintf(`You are generating cases for a Papers, Please-style document inspection game.

TODAY'S GAME DATE: %s

TODAY'S RULES:
- %s

Generate %d NPC worker cases. Each worker has TWO documents:
1. EMPLOYEE BADGE: name, picture (MUST be exactly "USE_CASE_ID_AS_SEED"), job_title, issue_date, expire_date, company_name
2. CLEARANCE FORM: name, shift_status (one of: "COMPLETE", "INCOMPLETE", "OVERTIME"), cargo1, cargo2

SHIFT STATUS MUST BE CLEAR:
- "COMPLETE" = Worker finished shift (can board if other rules pass)
- "INCOMPLETE" = Shift not finished (violation if rules require complete)
- "OVERTIME" = Extra hours worked (can board if rules allow)

CARGO MUST BE CLEAR AND SPECIFIC:
- ALLOWED: "Personal clothing", "Family photos", "Toiletries", "Snacks", "Music player", "Books", "Personal tablet", "Personal effects"
- COMPANY PROPERTY (violation): "Delta-7 drill bit", "Company mining equipment", "Work helmet", "Safety vest", "Company radio", "Excavation tools"
- CONTRABAND (violation): "Ore samples", "Mineral specimens", "Asteroid fragments", "Unauthorized samples"
- DO NOT use ambiguous items like "Research equipment" or "Tools" - be SPECIFIC about whether personal or company

NAME GENERATION - CRITICAL RULES:
- Generate UNIQUE, DIVERSE names for EVERY worker - NO REPETITION across all cases
- Use REALISTIC international names from varied cultures (Asian, African, European, Latin American, Middle Eastern, etc.)
- BANNED names - NEVER use: "Elara", "Kael", "Zephyr", "Lyra", "Anya", "Priya Sharma", "Omar Hassan" (overused)
- Mix cultural backgrounds: pair different ethnic first names with different ethnic last names
- Use uncommon but realistic combinations to ensure variety
- Think of actual real-world names you rarely see together
- EVERY case MUST have a completely different name from all others

Each case should have:
1. An NPC profile (name, role, department, personality, demeanor)
2. The two documents above
3. An opening line the NPC says
4. The ground truth about this worker
5. Whether they should be approved or denied
6. List of contradictions (if any) between documents

About 60%% should be APPROVED (compliant with rules).
About 40%% should be DENIED (violate at least one rule).

IMPORTANT DATE LOGIC:
- Today's date is %s
- Badge issue_date should be BEFORE today (e.g., 6 months ago)
- Badge expire_date can be AFTER today (valid) or BEFORE today (expired - violation!)
- Use the game date context to generate realistic dates

Return ONLY valid JSON with this exact structure:
{
  "cases": [
    {
      "npc": {
        "name": "Carlos Mendez",
        "role": "Mining Engineer",
        "department": "Excavation",
        "personality": "tired",
        "demeanor": "cooperative"
      },
      "documents": {
        "employee_badge": {
          "name": "Carlos Mendez",
          "picture": "USE_CASE_ID_AS_SEED",
          "job_title": "Mining Engineer",
          "issue_date": "YYYY-MM-DD (must be before today)",
          "expire_date": "YYYY-MM-DD (after today if valid, before if violation)",
          "company_name": "Delta-7 Mining Corp"
        },
        "clearance_form": {
          "name": "John Smith",
          "shift_status": "COMPLETE",
          "cargo1": "Personal effects",
          "cargo2": "None"
        }
      },
      "opening_line": "Hey, I need to catch the shuttle home. My shift's done.",
      "truth": {
        "employee_id": "EMP-1234",
        "should_approve": true,
        "reason": "Shift complete, badge valid, cargo approved"
      },
      "contradictions": [],
      "correct_decision": "approve"
    }
  ]
}`, gameDate, rulesText, count, gameDate)

	genConfig := &genai.GenerateContentConfig{
		Temperature: ptr(float32(1.0)),
	}

	resp, err := c.client.Models.GenerateContent(ctx, c.model, genai.Text(prompt), genConfig)
	if err != nil {
		return c.mockCases(count, gameDate), nil
	}

	text := resp.Text()
	if text == "" {
		return c.mockCases(count, gameDate), nil
	}

	// Extract JSON from response
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```json") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
	}

	var response struct {
		Cases []struct {
			NPC struct {
				Name        string `json:"name"`
				Role        string `json:"role"`
				Department  string `json:"department"`
				Personality string `json:"personality"`
				Demeanor    string `json:"demeanor"`
			} `json:"npc"`
			Documents struct {
				EmployeeBadge  map[string]string `json:"employee_badge"`
				ClearanceForm  map[string]string `json:"clearance_form"`
			} `json:"documents"`
			OpeningLine     string   `json:"opening_line"`
			Truth           struct {
				EmployeeID    string `json:"employee_id"`
				ShouldApprove bool   `json:"should_approve"`
				Reason        string `json:"reason"`
			} `json:"truth"`
			Contradictions  []string `json:"contradictions"`
			CorrectDecision string   `json:"correct_decision"`
		} `json:"cases"`
	}

	if err := json.Unmarshal([]byte(text), &response); err != nil {
		return c.mockCases(count, gameDate), nil
	}

	// Convert to models.Case
	cases := make([]models.Case, len(response.Cases))
	for i, geminiCase := range response.Cases {
		caseID := fmt.Sprintf("case-%d", i+1)

		// Fix badge picture URL to use caseID as seed
		badgeFields := geminiCase.Documents.EmployeeBadge
		if picture, ok := badgeFields["picture"]; ok && picture == "USE_CASE_ID_AS_SEED" {
			badgeFields["picture"] = fmt.Sprintf("https://api.dicebear.com/7.x/bottts/svg?seed=%s&backgroundColor=1a3a52&scale=90", caseID)
		}

		cases[i] = models.Case{
			CaseID: caseID,
			NPC: models.NPCProfile{
				Name:        geminiCase.NPC.Name,
				Role:        geminiCase.NPC.Role,
				Department:  geminiCase.NPC.Department,
				Personality: geminiCase.NPC.Personality,
				VoiceID:     "21m00Tcm4TlvDq8ikWAM",
				Demeanor:    geminiCase.NPC.Demeanor,
			},
			Documents: []models.Document{
				{Type: "employee_badge", Fields: badgeFields},
				{Type: "clearance_form", Fields: geminiCase.Documents.ClearanceForm},
			},
			OpeningLine: geminiCase.OpeningLine,
			Truth: models.CaseTruth{
				EmployeeID:    geminiCase.Truth.EmployeeID,
				ShouldApprove: geminiCase.Truth.ShouldApprove,
				Reason:        geminiCase.Truth.Reason,
			},
			Contradictions:  geminiCase.Contradictions,
			CorrectDecision: geminiCase.CorrectDecision,
		}
	}

	return cases, nil
}

// GenerateDialogue generates NPC response to player question
func (c *Client) GenerateDialogue(ctx context.Context, dialogueCtx models.DialogueContext) (string, error) {
	if err := c.initClient(ctx); err != nil {
		return "", err
	}

	// Fallback to mock if no client
	if c.client == nil {
		return fmt.Sprintf("I understand your question about '%s'. Let me explain...", dialogueCtx.Question), nil
	}

	prompt := fmt.Sprintf(`You are roleplaying an NPC worker at an asteroid mining station trying to board the final departure shuttle.

YOUR CHARACTER:
- Name: (implied from context)
- Role: %s
- Department: %s
- Personality: %s
- Demeanor: %s

THE TRUTH (player doesn't know this):
- Employee ID: %s
- Should be approved: %t
- Reason: %s

THE PLAYER ASKED: "%s"

Respond in character with 1-2 sentences. Be consistent with your personality and demeanor.
If the question reveals information that would expose contradictions, be slightly evasive or defensive.
If asked about something matching your documents, answer confidently.
Never break character. Never mention "the truth" explicitly.

Your response:`,
		dialogueCtx.NPCProfile.Role,
		dialogueCtx.NPCProfile.Department,
		dialogueCtx.NPCProfile.Personality,
		dialogueCtx.NPCProfile.Demeanor,
		dialogueCtx.CaseTruth.EmployeeID,
		dialogueCtx.CaseTruth.ShouldApprove,
		dialogueCtx.CaseTruth.Reason,
		dialogueCtx.Question)

	genConfig := &genai.GenerateContentConfig{
		Temperature: ptr(float32(1.2)),
	}

	resp, err := c.client.Models.GenerateContent(ctx, c.model, genai.Text(prompt), genConfig)
	if err != nil {
		return "I... uh... what was the question again?", nil
	}

	text := resp.Text()
	if text == "" {
		return "I'd rather not talk about that.", nil
	}

	return strings.TrimSpace(text), nil
}

// GenerateVerdict generates explanation of case outcome
func (c *Client) GenerateVerdict(ctx context.Context, caseData models.Case, playerDecision string) (string, error) {
	if err := c.initClient(ctx); err != nil {
		return "", err
	}

	correct := (playerDecision == caseData.CorrectDecision)

	// Fallback to mock if no client
	if c.client == nil {
		if correct {
			return fmt.Sprintf("Correct! %s", caseData.Truth.Reason), nil
		}
		return fmt.Sprintf("Incorrect. You should have %s. %s", caseData.CorrectDecision, caseData.Truth.Reason), nil
	}

	contradictions := "None"
	if len(caseData.Contradictions) > 0 {
		contradictions = strings.Join(caseData.Contradictions, "; ")
	}

	prompt := fmt.Sprintf(`You are a transit supervisor evaluating a clerk's document inspection decision in a Papers, Please-style game.

CASE DETAILS:
- Worker: %s (%s)
- Correct decision: %s
- Clerk's decision: %s
- Ground truth: %s
- Contradictions: %s

IMPORTANT - PERSPECTIVE:
- Address the PLAYER (the clerk making the decision)
- Provide feedback on the CLERK'S performance, NOT the worker's
- DO NOT say things like "Worker good job" or praise the worker
- DO say things like "Correct decision, clerk!" or "Wrong call!"

Generate a 1-2 sentence verdict addressing the clerk.

If correct: Praise the clerk and explain what they correctly identified (e.g., "Good catch, clerk! You correctly spotted the expired badge.")
If incorrect: Tell the clerk they made a mistake and explain what they missed (e.g., "Wrong decision! You failed to notice their shift status was INCOMPLETE.")

Be concise and professional like a transit supervisor evaluating their clerk.

Your verdict:`,
		caseData.NPC.Name,
		caseData.NPC.Role,
		caseData.CorrectDecision,
		playerDecision,
		caseData.Truth.Reason,
		contradictions)

	genConfig := &genai.GenerateContentConfig{
		Temperature: ptr(float32(0.7)),
	}

	resp, err := c.client.Models.GenerateContent(ctx, c.model, genai.Text(prompt), genConfig)
	if err != nil {
		if correct {
			return fmt.Sprintf("Correct! %s", caseData.Truth.Reason), nil
		}
		return fmt.Sprintf("Incorrect. You should have %s. %s", caseData.CorrectDecision, caseData.Truth.Reason), nil
	}

	text := resp.Text()
	if text == "" {
		if correct {
			return fmt.Sprintf("Correct! %s", caseData.Truth.Reason), nil
		}
		return fmt.Sprintf("Incorrect. You should have %s. %s", caseData.CorrectDecision, caseData.Truth.Reason), nil
	}

	return strings.TrimSpace(text), nil
}

// Mock data functions (fallbacks)

func (c *Client) mockRules() []string {
	// Define pools of possible rules
	shiftRules := []string{
		"Only COMPLETE shifts can board",
		"INCOMPLETE shifts stay on station",
		"OVERTIME workers need manager approval",
		"All shifts must be COMPLETE to depart",
	}

	cargoRules := []string{
		"No company equipment leaves the station",
		"Personal items only - no ore samples",
		"No company tools allowed on shuttle",
		"Contraband items = instant denial",
		"Only personal belongings permitted",
	}

	badgeRules := []string{
		"Expired badges get denied",
		"Badge must be valid on departure date",
		"No exceptions for expired credentials",
		"Current badges only - check dates",
	}

	// Randomly select rules from each category
	rules := []string{}

	// Pick 1-2 shift rules
	rules = append(rules, shiftRules[rand.Intn(len(shiftRules))])
	if rand.Float32() < 0.5 {
		secondShift := shiftRules[rand.Intn(len(shiftRules))]
		if secondShift != rules[0] {
			rules = append(rules, secondShift)
		}
	}

	// Pick 1 cargo rule
	rules = append(rules, cargoRules[rand.Intn(len(cargoRules))])

	// Pick 1 badge rule
	rules = append(rules, badgeRules[rand.Intn(len(badgeRules))])

	return rules
}

func (c *Client) mockCases(count int, gameDate string) []models.Case {
	cases := make([]models.Case, count)
	for i := 0; i < count; i++ {
		cases[i] = c.generateMockCase(i+1, gameDate)
	}
	return cases
}

func (c *Client) generateMockCase(index int, gameDate string) models.Case {
	workerName := fmt.Sprintf("Worker %d", index)
	jobTitle := "Drill Operator"
	shiftStatus := "COMPLETE"
	cargo1 := "Personal clothing"
	cargo2 := "Family photos"
	shouldApprove := true
	reason := "Shift complete, badge valid, cargo approved"

	// Vary some details based on index for variety and create violations
	switch index % 6 {
	case 0:
		jobTitle = "Drill Operator"
		shiftStatus = "COMPLETE"
		cargo1 = "Personal clothing"
		cargo2 = "Snacks"
		shouldApprove = true
		reason = "Shift complete, badge valid, cargo approved"
	case 1:
		jobTitle = "Ore Processor"
		shiftStatus = "COMPLETE"
		cargo1 = "Personal tablet"
		cargo2 = "Books"
		shouldApprove = true
		reason = "Shift complete, badge valid, cargo approved"
	case 2:
		jobTitle = "Systems Tech"
		shiftStatus = "COMPLETE"
		cargo1 = "Delta-7 drill bit"
		cargo2 = "Personal effects"
		shouldApprove = false
		reason = "Company equipment not allowed off-station"
	case 3:
		jobTitle = "Geologist"
		shiftStatus = "COMPLETE"
		cargo1 = "Ore samples"
		cargo2 = "Personal clothing"
		shouldApprove = false
		reason = "Ore samples are contraband"
	case 4:
		jobTitle = "Safety Officer"
		shiftStatus = "INCOMPLETE"
		cargo1 = "Music player"
		cargo2 = "Toiletries"
		shouldApprove = false
		reason = "Shift incomplete - cannot board"
	case 5:
		jobTitle = "Maintenance Tech"
		shiftStatus = "OVERTIME"
		cargo1 = "Family photos"
		cargo2 = "Personal clothing"
		shouldApprove = true
		reason = "Overtime shift complete, cargo approved"
	}

	// Parse game date and calculate badge dates
	parsedDate, err := time.Parse("2006-01-02", gameDate)
	if err != nil {
		// Fallback to current date + 100 years if parsing fails
		parsedDate = time.Now().AddDate(100, 0, 0)
	}

	// Badge issued 6 months ago, expires in 6 months
	badgeIssueDate := parsedDate.AddDate(0, -6, 0).Format("2006-01-02")
	badgeExpireDate := parsedDate.AddDate(0, 6, 0).Format("2006-01-02")

	caseID := fmt.Sprintf("case-%d", index)

	return models.Case{
		CaseID: caseID,
		NPC: models.NPCProfile{
			Name:        workerName,
			Role:        jobTitle,
			Department:  "Mining Operations",
			Personality: "tired",
			VoiceID:     "21m00Tcm4TlvDq8ikWAM",
			Demeanor:    "cooperative",
		},
		Documents: []models.Document{
			{
				Type: "employee_badge",
				Fields: map[string]string{
					"name":         workerName,
					"picture":      fmt.Sprintf("https://api.dicebear.com/7.x/bottts/svg?seed=%s&backgroundColor=1a3a52&scale=90", caseID),
					"job_title":    jobTitle,
					"issue_date":   badgeIssueDate,
					"expire_date":  badgeExpireDate,
					"company_name": "Delta-7 Mining Corp",
				},
			},
			{
				Type: "clearance_form",
				Fields: map[string]string{
					"name":         workerName,
					"shift_status": shiftStatus,
					"cargo1":       cargo1,
					"cargo2":       cargo2,
				},
			},
		},
		OpeningLine: "Hey, I need to catch the shuttle home. My shift's done.",
		Truth: models.CaseTruth{
			EmployeeID:       fmt.Sprintf("EMP-%04d", index),
			ActualTermEnd:    parsedDate.Format("2006-01-02"),
			ActualClearance:  "A",
			HasIncidents:     false,
			HasDebriefIssues: false,
			ShouldApprove:    shouldApprove,
			Reason:           reason,
		},
		Contradictions:  []string{},
		CorrectDecision: func() string {
			if shouldApprove {
				return "approve"
			}
			return "deny"
		}(),
	}
}
