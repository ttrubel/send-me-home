package firestore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/ttrubel/send-me-home/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	sessionsCollection = "sessions"
)

// Client handles session storage with Firestore
type Client struct {
	client *firestore.Client
}

func NewClient(projectID string) (*Client, error) {
	ctx := context.Background()

	if projectID == "" {
		return nil, fmt.Errorf("GOOGLE_CLOUD_PROJECT is required for Firestore")
	}

	client, err := firestore.NewClientWithDatabase(ctx, projectID, "send-me-home")
	if err != nil {
		return nil, fmt.Errorf("failed to create Firestore client: %w", err)
	}

	return &Client{
		client: client,
	}, nil
}

// Close closes the Firestore client
func (c *Client) Close() error {
	return c.client.Close()
}

// SaveSession stores a session
func (c *Client) SaveSession(ctx context.Context, session *models.Session) error {
	_, err := c.client.Collection(sessionsCollection).Doc(session.SessionID).Set(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
}

// GetSession retrieves a session
func (c *Client) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	doc, err := c.client.Collection(sessionsCollection).Doc(sessionID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session models.Session
	if err := doc.DataTo(&session); err != nil {
		return nil, fmt.Errorf("failed to parse session data: %w", err)
	}

	return &session, nil
}

// UpdateSession updates an existing session
func (c *Client) UpdateSession(ctx context.Context, session *models.Session) error {
	// Check if session exists first
	_, err := c.GetSession(ctx, session.SessionID)
	if err != nil {
		return err
	}

	// Update the session
	_, err = c.client.Collection(sessionsCollection).Doc(session.SessionID).Set(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

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
	docRef := c.client.Collection(sessionsCollection).Doc(sessionID)

	err := c.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(docRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return fmt.Errorf("session not found: %s", sessionID)
			}
			return err
		}

		var session models.Session
		if err := doc.DataTo(&session); err != nil {
			return fmt.Errorf("failed to parse session data: %w", err)
		}

		// Always increment to track completed cases
		// CurrentCaseIndex will equal len(session.Cases) when all cases are done
		session.CurrentCaseIndex++

		return tx.Set(docRef, session)
	})

	if err != nil {
		return fmt.Errorf("failed to increment case index: %w", err)
	}

	return nil
}

// UpdateScore updates the session score
func (c *Client) UpdateScore(ctx context.Context, sessionID string, scoreDelta int, correct bool) error {
	docRef := c.client.Collection(sessionsCollection).Doc(sessionID)

	err := c.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(docRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return fmt.Errorf("session not found: %s", sessionID)
			}
			return err
		}

		var session models.Session
		if err := doc.DataTo(&session); err != nil {
			return fmt.Errorf("failed to parse session data: %w", err)
		}

		session.Score += scoreDelta

		if correct {
			session.CorrectDecisions++
		} else {
			session.IncorrectDecisions++
		}

		return tx.Set(docRef, session)
	})

	if err != nil {
		return fmt.Errorf("failed to update score: %w", err)
	}

	return nil
}

// UseSecondaryCheck decrements the secondary check quota
func (c *Client) UseSecondaryCheck(ctx context.Context, sessionID string) error {
	docRef := c.client.Collection(sessionsCollection).Doc(sessionID)

	err := c.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(docRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return fmt.Errorf("session not found: %s", sessionID)
			}
			return err
		}

		var session models.Session
		if err := doc.DataTo(&session); err != nil {
			return fmt.Errorf("failed to parse session data: %w", err)
		}

		if session.RemainingSecondaryChecks <= 0 {
			return fmt.Errorf("no secondary checks remaining")
		}

		session.RemainingSecondaryChecks--

		return tx.Set(docRef, session)
	})

	if err != nil {
		return fmt.Errorf("failed to use secondary check: %w", err)
	}

	return nil
}
