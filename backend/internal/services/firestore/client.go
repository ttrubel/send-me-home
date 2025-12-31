package firestore

import (
	"context"
	"fmt"
	"sync"

	"github.com/ttrubel/send-me-home/internal/models"
)

// Client handles session storage
// TODO: Replace in-memory storage with actual Firestore
type Client struct {
	sessions map[string]*models.Session
	mu       sync.RWMutex
}

func NewClient(projectID string) (*Client, error) {
	// TODO: Initialize Firestore client
	return &Client{
		sessions: make(map[string]*models.Session),
	}, nil
}

// SaveSession stores a session
func (c *Client) SaveSession(ctx context.Context, session *models.Session) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.sessions[session.SessionID] = session
	return nil
}

// GetSession retrieves a session
func (c *Client) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	session, ok := c.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// UpdateSession updates an existing session
func (c *Client) UpdateSession(ctx context.Context, session *models.Session) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.sessions[session.SessionID]; !ok {
		return fmt.Errorf("session not found: %s", session.SessionID)
	}

	c.sessions[session.SessionID] = session
	return nil
}

// GetCase retrieves a specific case from a session
func (c *Client) GetCase(ctx context.Context, sessionID, caseID string) (*models.Case, error) {
	session, err := c.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	for _, caseData := range session.Cases {
		if caseData.CaseID == caseID {
			return &caseData, nil
		}
	}

	return nil, fmt.Errorf("case not found: %s", caseID)
}

// IncrementCaseIndex moves to the next case
func (c *Client) IncrementCaseIndex(ctx context.Context, sessionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	session, ok := c.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Always increment to track completed cases
	// CurrentCaseIndex will equal len(session.Cases) when all cases are done
	session.CurrentCaseIndex++

	return nil
}

// UpdateScore updates the session score
func (c *Client) UpdateScore(ctx context.Context, sessionID string, scoreDelta int, correct bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	session, ok := c.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.Score += scoreDelta

	if correct {
		session.CorrectDecisions++
	} else {
		session.IncorrectDecisions++
	}

	return nil
}

// UseSecondaryCheck decrements the secondary check quota
func (c *Client) UseSecondaryCheck(ctx context.Context, sessionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	session, ok := c.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if session.RemainingSecondaryChecks <= 0 {
		return fmt.Errorf("no secondary checks remaining")
	}

	session.RemainingSecondaryChecks--
	return nil
}
