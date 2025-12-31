package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
func (c *Client) GenerateRules(ctx context.Context) ([]string, error) {
	if err := c.initClient(ctx); err != nil {
		return nil, err
	}

	// Fallback to mock if no client
	if c.client == nil {
		return c.mockRules(), nil
	}

	prompt := `You are generating rules for a Papers, Please-style game set on an asteroid mining station.

Generate 3-5 daily transit rules that workers must comply with to board the final departure shuttle.

Rules should:
- Be specific and verifiable from documents (contract, shift log, clearance badge)
- Create interesting edge cases and contradictions
- Sound like bureaucratic regulations
- Relate to: contract completion, incidents, medical clearance, fees, zone access

Return ONLY a JSON array of strings, no other text:
["rule 1", "rule 2", "rule 3"]`

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
func (c *Client) GenerateCases(ctx context.Context, rules []string, count int) ([]models.Case, error) {
	if err := c.initClient(ctx); err != nil {
		return nil, err
	}

	// Fallback to mock if no client
	if c.client == nil {
		return c.mockCases(count), nil
	}

	rulesText := strings.Join(rules, "\n- ")

	prompt := fmt.Sprintf(`You are generating cases for a Papers, Please-style document inspection game.

TODAY'S RULES:
- %s

Generate %d NPC worker cases. Each case should have:
1. An NPC profile (name, role, department, personality, demeanor)
2. Three documents: contract, shift_log, clearance_badge
3. An opening line the NPC says
4. The ground truth about this worker
5. Whether they should be approved or denied
6. List of contradictions (if any) between documents

About 60%% should be APPROVED (compliant with rules).
About 40%% should be DENIED (violate at least one rule).

Return ONLY valid JSON with this exact structure:
{
  "cases": [
    {
      "npc": {
        "name": "John Smith",
        "role": "Mining Engineer",
        "department": "Excavation",
        "personality": "tired",
        "demeanor": "cooperative"
      },
      "documents": {
        "contract": {
          "name": "John Smith",
          "employee_id": "EMP-1234",
          "role": "Mining Engineer",
          "department": "Excavation",
          "term_end": "2024-12-25",
          "signature": "✓ Signed"
        },
        "shift_log": {
          "employee_id": "EMP-1234",
          "last_shift": "2024-12-24",
          "total_hours": "2080",
          "incidents": "None",
          "debrief_status": "Complete"
        },
        "clearance_badge": {
          "employee_id": "EMP-1234",
          "access_level": "A",
          "medical_clearance": "Valid",
          "zone_auth": "Zone A, B",
          "expires": "2025-01-15"
        }
      },
      "opening_line": "Hey, I need to catch the shuttle home. My contract's up.",
      "truth": {
        "employee_id": "EMP-1234",
        "should_approve": true,
        "reason": "Contract term complete, no incidents, valid clearance"
      },
      "contradictions": [],
      "correct_decision": "approve"
    }
  ]
}`, rulesText, count)

	genConfig := &genai.GenerateContentConfig{
		Temperature: ptr(float32(1.0)),
	}

	resp, err := c.client.Models.GenerateContent(ctx, c.model, genai.Text(prompt), genConfig)
	if err != nil {
		return c.mockCases(count), nil
	}

	text := resp.Text()
	if text == "" {
		return c.mockCases(count), nil
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
				Contract       map[string]string `json:"contract"`
				ShiftLog       map[string]string `json:"shift_log"`
				ClearanceBadge map[string]string `json:"clearance_badge"`
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
		return c.mockCases(count), nil
	}

	// Convert to models.Case
	cases := make([]models.Case, len(response.Cases))
	for i, geminiCase := range response.Cases {
		cases[i] = models.Case{
			CaseID: fmt.Sprintf("case-%d", i+1),
			NPC: models.NPCProfile{
				Name:        geminiCase.NPC.Name,
				Role:        geminiCase.NPC.Role,
				Department:  geminiCase.NPC.Department,
				Personality: geminiCase.NPC.Personality,
				VoiceID:     "21m00Tcm4TlvDq8ikWAM",
				Demeanor:    geminiCase.NPC.Demeanor,
			},
			Documents: []models.Document{
				{Type: "contract", Fields: geminiCase.Documents.Contract},
				{Type: "shift_log", Fields: geminiCase.Documents.ShiftLog},
				{Type: "clearance_badge", Fields: geminiCase.Documents.ClearanceBadge},
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

	prompt := fmt.Sprintf(`You are explaining the outcome of a document inspection decision in a Papers, Please-style game.

CASE DETAILS:
- Worker: %s (%s)
- Correct decision: %s
- Player decision: %s
- Ground truth: %s
- Contradictions: %s

Generate a 1-2 sentence explanation of why the player was correct or incorrect.

If correct: Praise briefly and explain what they caught.
If incorrect: Explain what they missed and what the correct decision was.

Be concise and professional like a transit supervisor.

Your explanation:`,
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
	return []string{
		"Only workers with contract term complete may board",
		"Any incident flag requires supervisor sign-off",
		"Medical clearance required after exposure events",
		"No unpaid equipment fees allowed",
		"All personnel must have valid Zone A clearance",
	}
}

func (c *Client) mockCases(count int) []models.Case {
	cases := make([]models.Case, count)
	for i := 0; i < count; i++ {
		cases[i] = c.generateMockCase(i + 1)
	}
	return cases
}

func (c *Client) generateMockCase(index int) models.Case {
	employeeID := fmt.Sprintf("EMP-%04d", index)

	return models.Case{
		CaseID: fmt.Sprintf("case-%d", index),
		NPC: models.NPCProfile{
			Name:        fmt.Sprintf("Worker %d", index),
			Role:        "Drill Operator",
			Department:  "Mining Operations",
			Personality: "tired",
			VoiceID:     "21m00Tcm4TlvDq8ikWAM",
			Demeanor:    "cooperative",
		},
		Documents: []models.Document{
			{
				Type: "contract",
				Fields: map[string]string{
					"name":        fmt.Sprintf("Worker %d", index),
					"employee_id": employeeID,
					"role":        "Drill Operator",
					"department":  "Mining Operations",
					"term_end":    "2024-12-25",
					"signature":   "✓ Signed",
				},
			},
			{
				Type: "shift_log",
				Fields: map[string]string{
					"employee_id":    employeeID,
					"last_shift":     "2024-12-24",
					"total_hours":    "2,080",
					"incidents":      "None",
					"debrief_status": "Complete",
				},
			},
			{
				Type: "clearance_badge",
				Fields: map[string]string{
					"employee_id":       employeeID,
					"access_level":      "A",
					"medical_clearance": "Valid",
					"zone_auth":         "Zone A, B",
					"expires":           "2025-01-15",
				},
			},
		},
		OpeningLine: "Hey, I need to catch the shuttle home. My contract's up.",
		Truth: models.CaseTruth{
			EmployeeID:       employeeID,
			ActualTermEnd:    "2024-12-25",
			ActualClearance:  "A",
			HasIncidents:     false,
			HasDebriefIssues: false,
			ShouldApprove:    true,
			Reason:           "Contract term complete, no incidents, valid clearance",
		},
		Contradictions:  []string{},
		CorrectDecision: "approve",
	}
}
