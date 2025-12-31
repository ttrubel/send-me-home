package api

import (
	"context"
	"fmt"
	"log"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	gamev1 "github.com/ttrubel/send-me-home/gen/game/v1"
	"github.com/ttrubel/send-me-home/internal/models"
	"github.com/ttrubel/send-me-home/internal/services/elevenlabs"
	"github.com/ttrubel/send-me-home/internal/services/firestore"
	"github.com/ttrubel/send-me-home/internal/services/gemini"
)

type GameHandler struct {
	gemini     *gemini.Client
	firestore  *firestore.Client
	elevenlabs *elevenlabs.Client
}

func NewGameHandler(geminiClient *gemini.Client, firestoreClient *firestore.Client, elevenlabsClient *elevenlabs.Client) *GameHandler {
	return &GameHandler{
		gemini:     geminiClient,
		firestore:  firestoreClient,
		elevenlabs: elevenlabsClient,
	}
}

// StartSession generates all cases upfront and returns progress updates
func (h *GameHandler) StartSession(
	ctx context.Context,
	req *connect.Request[gamev1.StartSessionRequest],
	stream *connect.ServerStream[gamev1.StartSessionResponse],
) error {
	numCases := int(req.Msg.NumCases)
	if numCases <= 0 {
		numCases = 15
	}

	// Set game date to current date + 100 years
	gameDate := time.Now().AddDate(100, 0, 0).Format("2006-01-02")

	// Step 1: Generate rules
	stream.Send(&gamev1.StartSessionResponse{
		Update: &gamev1.StartSessionResponse_Progress{
			Progress: &gamev1.SessionProgress{
				Current: 0,
				Total:   int32(numCases),
				Message: "Generating daily rules...",
			},
		},
	})

	rules, err := h.gemini.GenerateRules(ctx, gameDate)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to generate rules: %w", err))
	}

	// Step 2: Generate cases
	for i := 0; i < numCases; i++ {
		stream.Send(&gamev1.StartSessionResponse{
			Update: &gamev1.StartSessionResponse_Progress{
				Progress: &gamev1.SessionProgress{
					Current: int32(i),
					Total:   int32(numCases),
					Message: fmt.Sprintf("Generating case %d/%d...", i+1, numCases),
				},
			},
		})
	}

	cases, err := h.gemini.GenerateCases(ctx, rules, numCases, gameDate)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to generate cases: %w", err))
	}

	// Step 2.5: Generate opening audio for each case with ElevenLabs
	for i := range cases {
		stream.Send(&gamev1.StartSessionResponse{
			Update: &gamev1.StartSessionResponse_Progress{
				Progress: &gamev1.SessionProgress{
					Current: int32(i),
					Total:   int32(numCases),
					Message: fmt.Sprintf("Generating voice audio %d/%d...", i+1, numCases),
				},
			},
		})

		// Generate audio for opening line
		audioData, err := h.elevenlabs.TextToSpeech(ctx, cases[i].NPC.VoiceID, cases[i].OpeningLine)
		if err != nil {
			log.Printf("Warning: Failed to generate audio for case %d: %v", i, err)
			// Continue without audio - it's optional
		} else if audioData != nil {
			cases[i].OpeningAudio = audioData
		}
	}

	// Step 3: Create session
	sessionID := uuid.New().String()
	secondaryChecksQuota := 3

	session := &models.Session{
		SessionID:                sessionID,
		GameDate:                 gameDate,
		Rules:                    rules,
		Cases:                    cases,
		CurrentCaseIndex:         0,
		Score:                    0,
		CorrectDecisions:         0,
		IncorrectDecisions:       0,
		SecondaryChecksQuota:     secondaryChecksQuota,
		RemainingSecondaryChecks: secondaryChecksQuota,
		CompletedCases:           []string{},
	}

	if err := h.firestore.SaveSession(ctx, session); err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to save session: %w", err))
	}

	// Step 4: Send ready signal
	stream.Send(&gamev1.StartSessionResponse{
		Update: &gamev1.StartSessionResponse_Ready{
			Ready: &gamev1.SessionReady{
				SessionId:            sessionID,
				GameDate:             gameDate,
				Rules:                rules,
				TotalCases:           int32(numCases),
				SecondaryChecksQuota: int32(secondaryChecksQuota),
			},
		},
	})

	log.Printf("Session started: %s with %d cases", sessionID, numCases)
	return nil
}

// GetNextCase returns the next pre-generated case
func (h *GameHandler) GetNextCase(
	ctx context.Context,
	req *connect.Request[gamev1.GetNextCaseRequest],
) (*connect.Response[gamev1.GetNextCaseResponse], error) {
	session, err := h.firestore.GetSession(ctx, req.Msg.SessionId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	if session.CurrentCaseIndex >= len(session.Cases) {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("no more cases available"))
	}

	currentCase := session.Cases[session.CurrentCaseIndex]

	// Convert models.Document to protobuf Document
	docs := make([]*gamev1.Document, len(currentCase.Documents))
	for i, doc := range currentCase.Documents {
		docs[i] = &gamev1.Document{
			Type:      doc.Type,
			Fields:    doc.Fields,
			VisualUrl: doc.VisualURL,
		}
	}

	response := &gamev1.GetNextCaseResponse{
		CaseId: currentCase.CaseID,
		Npc: &gamev1.NPCProfile{
			Name:        currentCase.NPC.Name,
			Role:        currentCase.NPC.Role,
			Department:  currentCase.NPC.Department,
			Personality: currentCase.NPC.Personality,
			PortraitUrl: currentCase.NPC.PortraitURL,
			Demeanor:    currentCase.NPC.Demeanor,
		},
		Documents:                docs,
		OpeningLine:              currentCase.OpeningLine,
		OpeningAudio:             currentCase.OpeningAudio,
		CaseNumber:               int32(session.CurrentCaseIndex + 1),
		RemainingSecondaryChecks: int32(session.RemainingSecondaryChecks),
	}

	return connect.NewResponse(response), nil
}

// AskQuestion generates NPC response to player question
func (h *GameHandler) AskQuestion(
	ctx context.Context,
	req *connect.Request[gamev1.AskQuestionRequest],
	stream *connect.ServerStream[gamev1.AskQuestionResponse],
) error {
	caseData, err := h.firestore.GetCase(ctx, req.Msg.SessionId, req.Msg.CaseId)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, err)
	}

	// Generate dialogue with Gemini
	dialogueCtx := models.DialogueContext{
		Question:   req.Msg.Question,
		CaseTruth:  caseData.Truth,
		NPCProfile: caseData.NPC,
	}

	responseText, err := h.gemini.GenerateDialogue(ctx, dialogueCtx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}

	// Send text chunk
	stream.Send(&gamev1.AskQuestionResponse{
		Chunk: &gamev1.AskQuestionResponse_TextChunk{
			TextChunk: responseText,
		},
	})

	// Generate and stream audio with ElevenLabs
	audioData, err := h.elevenlabs.TextToSpeech(ctx, caseData.NPC.VoiceID, responseText)
	if err != nil {
		log.Printf("Warning: Failed to generate audio for response: %v", err)
		// Continue without audio - it's optional
	} else if audioData != nil {
		// Send audio chunk
		stream.Send(&gamev1.AskQuestionResponse{
			Chunk: &gamev1.AskQuestionResponse_AudioChunk{
				AudioChunk: audioData,
			},
		})
	}

	// Send done signal
	stream.Send(&gamev1.AskQuestionResponse{
		Chunk: &gamev1.AskQuestionResponse_Done{
			Done: true,
		},
	})

	return nil
}

// SecondaryCheck performs verification check
func (h *GameHandler) SecondaryCheck(
	ctx context.Context,
	req *connect.Request[gamev1.SecondaryCheckRequest],
) (*connect.Response[gamev1.SecondaryCheckResponse], error) {
	session, err := h.firestore.GetSession(ctx, req.Msg.SessionId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	if session.RemainingSecondaryChecks <= 0 {
		return nil, connect.NewError(connect.CodeResourceExhausted, fmt.Errorf("no secondary checks remaining"))
	}

	caseData, err := h.firestore.GetCase(ctx, req.Msg.SessionId, req.Msg.CaseId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	// Use a secondary check
	if err := h.firestore.UseSecondaryCheck(ctx, req.Msg.SessionId); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Verify employee ID against truth
	valid := (req.Msg.EmployeeId == caseData.Truth.EmployeeID)
	message := fmt.Sprintf("Contract verified: Term ends %s", caseData.Truth.ActualTermEnd)

	response := &gamev1.SecondaryCheckResponse{
		Valid:           valid,
		Message:         message,
		RemainingChecks: int32(session.RemainingSecondaryChecks - 1),
	}

	return connect.NewResponse(response), nil
}

// ResolveCase handles player decision and returns verdict
func (h *GameHandler) ResolveCase(
	ctx context.Context,
	req *connect.Request[gamev1.ResolveCaseRequest],
) (*connect.Response[gamev1.ResolveCaseResponse], error) {
	session, err := h.firestore.GetSession(ctx, req.Msg.SessionId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	caseData, err := h.firestore.GetCase(ctx, req.Msg.SessionId, req.Msg.CaseId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	// Map decision to string
	playerDecision := ""
	switch req.Msg.Decision {
	case gamev1.Decision_DECISION_APPROVE:
		playerDecision = "approve"
	case gamev1.Decision_DECISION_DENY:
		playerDecision = "deny"
	case gamev1.Decision_DECISION_SECONDARY:
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("use SecondaryCheck endpoint instead"))
	}

	// Check if correct
	correct := (playerDecision == caseData.CorrectDecision)

	// Calculate score
	scoreDelta := 10
	if !correct {
		scoreDelta = -15
	}

	// Update session
	if err := h.firestore.UpdateScore(ctx, req.Msg.SessionId, scoreDelta, correct); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Move to next case
	if err := h.firestore.IncrementCaseIndex(ctx, req.Msg.SessionId); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Generate verdict
	verdict, err := h.gemini.GenerateVerdict(ctx, *caseData, playerDecision)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Determine outcome
	outcome := gamev1.CaseOutcome_CASE_OUTCOME_UNSPECIFIED
	if correct && playerDecision == "approve" {
		outcome = gamev1.CaseOutcome_CASE_OUTCOME_CORRECT_APPROVE
	} else if correct && playerDecision == "deny" {
		outcome = gamev1.CaseOutcome_CASE_OUTCOME_CORRECT_DENY
	} else if !correct && playerDecision == "approve" {
		outcome = gamev1.CaseOutcome_CASE_OUTCOME_WRONG_APPROVE
	} else if !correct && playerDecision == "deny" {
		outcome = gamev1.CaseOutcome_CASE_OUTCOME_WRONG_DENY
	}

	// Refresh session for updated score
	session, _ = h.firestore.GetSession(ctx, req.Msg.SessionId)

	response := &gamev1.ResolveCaseResponse{
		Correct:             correct,
		Verdict:             verdict,
		ContradictionsFound: caseData.Contradictions,
		ScoreDelta:          int32(scoreDelta),
		TotalScore:          int32(session.Score),
		Outcome:             outcome,
	}

	return connect.NewResponse(response), nil
}

// GetSessionStatus returns current session stats
func (h *GameHandler) GetSessionStatus(
	ctx context.Context,
	req *connect.Request[gamev1.GetSessionStatusRequest],
) (*connect.Response[gamev1.GetSessionStatusResponse], error) {
	session, err := h.firestore.GetSession(ctx, req.Msg.SessionId)
	if err != nil {
		log.Printf("ERROR GetSessionStatus: Session not found: %s, error: %v", req.Msg.SessionId, err)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	log.Printf("GetSessionStatus called for session %s:", req.Msg.SessionId)
	log.Printf("  CurrentCaseIndex: %d", session.CurrentCaseIndex)
	log.Printf("  Total Cases: %d", len(session.Cases))
	log.Printf("  Score: %d", session.Score)
	log.Printf("  Correct: %d", session.CorrectDecisions)
	log.Printf("  Incorrect: %d", session.IncorrectDecisions)

	response := &gamev1.GetSessionStatusResponse{
		CasesCompleted:           int32(session.CurrentCaseIndex),
		TotalCases:               int32(len(session.Cases)),
		TotalScore:               int32(session.Score),
		CorrectDecisions:         int32(session.CorrectDecisions),
		IncorrectDecisions:       int32(session.IncorrectDecisions),
		RemainingSecondaryChecks: int32(session.RemainingSecondaryChecks),
		SessionComplete:          session.CurrentCaseIndex >= len(session.Cases),
	}

	return connect.NewResponse(response), nil
}
