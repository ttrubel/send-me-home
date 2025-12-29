package gemini

import (
	"context"
	"fmt"

	"github.com/ttrubel/send-me-home/internal/models"
)

type Client struct {
	apiKey string
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
	}
}

// GenerateRules generates daily rules for the shift
func (c *Client) GenerateRules(ctx context.Context) ([]string, error) {
	// TODO: Implement Gemini API call to generate rules
	// For now, return mock rules
	return []string{
		"Only workers with contract term complete may board",
		"Any incident flag requires supervisor sign-off",
		"Medical clearance required after exposure events",
		"No unpaid equipment fees allowed",
		"All personnel must have valid Zone A clearance",
	}, nil
}

// GenerateCases generates multiple cases in parallel
func (c *Client) GenerateCases(ctx context.Context, rules []string, count int) ([]models.Case, error) {
	// TODO: Implement parallel case generation with Gemini API
	// For now, return mock cases
	cases := make([]models.Case, count)

	for i := 0; i < count; i++ {
		cases[i] = c.generateMockCase(i + 1)
	}

	return cases, nil
}

// GenerateDialogue generates NPC response to player question
func (c *Client) GenerateDialogue(ctx context.Context, dialogueCtx models.DialogueContext) (string, error) {
	// TODO: Implement Gemini API call for dialogue generation
	// For now, return mock response
	return fmt.Sprintf("I understand your question about '%s'. Let me explain...", dialogueCtx.Question), nil
}

// GenerateVerdict generates explanation of case outcome
func (c *Client) GenerateVerdict(ctx context.Context, caseData models.Case, playerDecision string) (string, error) {
	// TODO: Implement Gemini API call for verdict generation
	correct := (playerDecision == caseData.CorrectDecision)

	if correct {
		return fmt.Sprintf("Correct! %s", caseData.Truth.Reason), nil
	}

	return fmt.Sprintf("Incorrect. You should have %s. %s", caseData.CorrectDecision, caseData.Truth.Reason), nil
}

// generateMockCase creates a mock case for testing
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
