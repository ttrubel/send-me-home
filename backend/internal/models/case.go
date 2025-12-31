package models

// Case represents a pre-generated NPC case
type Case struct {
	CaseID       string                 `json:"case_id"`
	NPC          NPCProfile             `json:"npc"`
	Documents    []Document             `json:"documents"`
	OpeningLine  string                 `json:"opening_line"`
	OpeningAudio []byte                 `json:"opening_audio,omitempty"`
	Truth        CaseTruth              `json:"truth"`
	Contradictions []string             `json:"contradictions"`
	CorrectDecision string              `json:"correct_decision"` // "approve" or "deny"
}

// NPCProfile contains NPC personality and appearance
type NPCProfile struct {
	Name        string `json:"name"`
	Role        string `json:"role"`
	Department  string `json:"department"`
	Personality string `json:"personality"` // "tired", "angry", "nervous"
	VoiceID     string `json:"voice_id"`
	Demeanor    string `json:"demeanor"` // "evasive", "cooperative", "frustrated"
	PortraitURL string `json:"portrait_url,omitempty"`
}

// Document represents a game document
type Document struct {
	Type      string            `json:"type"` // "contract", "shift_log", "clearance_badge"
	Fields    map[string]string `json:"fields"`
	VisualURL string            `json:"visual_url,omitempty"`
}

// CaseTruth contains the hidden ground truth for a case
type CaseTruth struct {
	EmployeeID       string `json:"employee_id"`
	ActualTermEnd    string `json:"actual_term_end"`
	ActualClearance  string `json:"actual_clearance"`
	HasIncidents     bool   `json:"has_incidents"`
	HasDebriefIssues bool   `json:"has_debrief_issues"`
	ShouldApprove    bool   `json:"should_approve"`
	Reason           string `json:"reason"`
}

// Session represents a game session
type Session struct {
	SessionID              string   `json:"session_id"`
	GameDate               string   `json:"game_date"` // Current game date (e.g. "2084-12-25")
	Rules                  []string `json:"rules"`
	Cases                  []Case   `json:"cases"`
	CurrentCaseIndex       int      `json:"current_case_index"`
	Score                  int      `json:"score"`
	CorrectDecisions       int      `json:"correct_decisions"`
	IncorrectDecisions     int      `json:"incorrect_decisions"`
	SecondaryChecksQuota   int      `json:"secondary_checks_quota"`
	RemainingSecondaryChecks int    `json:"remaining_secondary_checks"`
	CompletedCases         []string `json:"completed_cases"`
}

// DialogueContext holds context for generating NPC responses
type DialogueContext struct {
	Question    string     `json:"question"`
	CaseTruth   CaseTruth  `json:"case_truth"`
	NPCProfile  NPCProfile `json:"npc_profile"`
	AskedQuestions []string `json:"asked_questions"`
}
