import { useEffect, useRef, useState } from 'react';
import { gameClient } from '../api/client';
import { audioManager } from '../audio/AudioManager';
import type { Document, NPCProfile } from '../gen/game/v1/game_pb';
import { Decision } from '../gen/game/v1/game_pb';
import './GameDesk.css';

interface GameDeskProps {
  sessionId: string;
  gameDate: string;
  rules: string[];
  totalCases: number;
  onComplete: (stats: {
    totalScore: number;
    correct: number;
    incorrect: number;
    accuracy: number;
  }) => void;
}

interface CaseData {
  caseId: string;
  npc: NPCProfile;
  documents: Document[];
  openingLine: string;
  caseNumber: number;
  remainingSecondaryChecks: number;
}

function GameDesk({ sessionId, gameDate, rules, totalCases, onComplete }: GameDeskProps) {
  const [currentCase, setCurrentCase] = useState<CaseData | null>(null);
  const [score, setScore] = useState(0);
  const [question, setQuestion] = useState('');
  const [npcResponse, setNpcResponse] = useState('');
  const [verdict, setVerdict] = useState('');
  const [showVerdict, setShowVerdict] = useState(false);
  const [loading, setLoading] = useState(false);
  const questionInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    loadNextCase();
  }, []);

  // Space bar handler for quick access to question input
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.code === 'Space' && e.target === document.body && !showVerdict) {
        e.preventDefault();
        questionInputRef.current?.focus();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [showVerdict]);

  const loadNextCase = async () => {
    setLoading(true);
    setShowVerdict(false);
    setVerdict('');
    setNpcResponse('');

    try {
      const response: any = await gameClient.getNextCase({ sessionId });

      setCurrentCase({
        caseId: response.caseId,
        npc: response.npc!,
        documents: response.documents,
        openingLine: response.openingLine,
        caseNumber: response.caseNumber,
        remainingSecondaryChecks: response.remainingSecondaryChecks,
      });

      audioManager.playNewCaseSound();

      // Play opening audio if available
      if (response.openingAudio && response.openingAudio.length > 0) {
        const audioBlob = new Blob([response.openingAudio], { type: 'audio/mpeg' });
        const audioUrl = URL.createObjectURL(audioBlob);
        const audio = new Audio(audioUrl);
        audio.play().catch((err) => console.error('Failed to play opening audio:', err));
      }
    } catch (error: any) {
      if (error.message?.includes('no more cases')) {
        // Fetch final session stats
        try {
          const statusResponse: any = await gameClient.getSessionStatus({ sessionId });
          const accuracy = statusResponse.casesCompleted > 0
            ? Math.round((statusResponse.correctDecisions / statusResponse.casesCompleted) * 100)
            : 0;

          onComplete({
            totalScore: statusResponse.totalScore,
            correct: statusResponse.correctDecisions,
            incorrect: statusResponse.incorrectDecisions,
            accuracy,
          });
        } catch (statsError) {
          console.error('Failed to fetch session stats:', statsError);
          console.error('Session ID:', sessionId);
          console.error('Error details:', JSON.stringify(statsError, null, 2));
          alert('ERROR: Failed to fetch final stats. Check console for details.');
          // Fallback with current score
          onComplete({
            totalScore: score,
            correct: 0,
            incorrect: 0,
            accuracy: 0,
          });
        }
      } else {
        console.error('Failed to load case:', error);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleAskQuestion = async () => {
    if (!question.trim() || !currentCase) return;

    audioManager.playAskQuestionSound();

    setLoading(true);
    setNpcResponse('');

    try {
      const stream: any = gameClient.askQuestion({
        sessionId,
        caseId: currentCase.caseId,
        question,
      });

      for await (const response of stream) {
        if (response.chunk.case === 'textChunk') {
          setNpcResponse(response.chunk.value);
        } else if (response.chunk.case === 'audioChunk') {
          // Play audio chunk
          const audioBlob = new Blob([response.chunk.value], { type: 'audio/mpeg' });
          const audioUrl = URL.createObjectURL(audioBlob);
          const audio = new Audio(audioUrl);
          audio.play().catch((err) => console.error('Failed to play response audio:', err));
        }
      }
    } catch (error) {
      console.error('Failed to ask question:', error);
    } finally {
      setLoading(false);
      setQuestion('');
    }
  };

  const handleDecision = async (decision: Decision) => {
    if (!currentCase) return;

    if (decision === Decision.APPROVE) {
      audioManager.playApproveSound();
    } else if (decision === Decision.DENY) {
      audioManager.playDenySound();
    }

    setLoading(true);

    try {
      const response: any = await gameClient.resolveCase({
        sessionId,
        caseId: currentCase.caseId,
        decision,
      });

      // Play NPC reaction audio if available
      if (response.npcReactionAudio && response.npcReactionAudio.length > 0) {
        const audioBlob = new Blob([response.npcReactionAudio], { type: 'audio/mpeg' });
        const audioUrl = URL.createObjectURL(audioBlob);
        const audio = new Audio(audioUrl);
        audio.play().catch((err) => console.error('Failed to play reaction audio:', err));
      }

      // Display NPC reaction text first, then verdict after a short delay
      if (response.npcReactionText) {
        setNpcResponse(response.npcReactionText);
        // Show verdict after a delay to let player see reaction
        await new Promise(resolve => setTimeout(resolve, 1500));
      }

      setVerdict(response.verdict);
      setScore(response.totalScore);
      setShowVerdict(true);
    } catch (error) {
      console.error('Failed to resolve case:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleNextCase = async () => {
    audioManager.playButtonClick();

    // Check if this was the last case
    if (currentCase && currentCase.caseNumber >= totalCases) {
      // Fetch final session stats and complete the game
      try {
        const statusResponse: any = await gameClient.getSessionStatus({ sessionId });
        const accuracy = statusResponse.casesCompleted > 0
          ? Math.round((statusResponse.correctDecisions / statusResponse.casesCompleted) * 100)
          : 0;

        onComplete({
          totalScore: statusResponse.totalScore,
          correct: statusResponse.correctDecisions,
          incorrect: statusResponse.incorrectDecisions,
          accuracy,
        });
      } catch (statsError) {
        console.error('Failed to fetch session stats (handleNextCase):', statsError);
        console.error('Session ID:', sessionId);
        console.error('Error details:', JSON.stringify(statsError, null, 2));
        alert('ERROR: Failed to fetch final stats. Check console for details.');
        // Fallback with current score
        onComplete({
          totalScore: score,
          correct: 0,
          incorrect: 0,
          accuracy: 0,
        });
      }
    } else {
      // Load next case
      loadNextCase();
    }
  };

  if (loading && !currentCase) {
    return <div className="loading">▮▮▮ LOADING NEXT CASE ▮▮▮</div>;
  }

  if (!currentCase) {
    return null;
  }

  return (
    <>
      <div className="game-desk">
        {/* TOP BAR - Game Stats */}
        <div className="top-bar">
          <div className="stat-display">
            <span className="stat-label">DATE:</span>
            <span className="stat-value">{gameDate}</span>
          </div>
          <div className="stat-display">
            <span className="stat-label">CASE:</span>
            <span className="stat-value">{currentCase.caseNumber}/{totalCases}</span>
          </div>
          <div className="stat-display">
            <span className="stat-label">SCORE:</span>
            <span className="stat-value score">{score}</span>
          </div>
          <div className="stat-display">
            <span className="stat-label">RULES:</span>
            <span className="stat-value">{rules.length} ACTIVE</span>
          </div>
        </div>

        {/* LEFT PANEL - Documents */}
        <div className="left-panel">
          <div className="panel-header">WORKER DOCUMENTS</div>
          <div className="documents-area">
            {currentCase.documents.map((doc, i) => (
              <div key={i} className="document">
                <div className="document-type">
                  {doc.type.replace('_', ' ')}
                </div>
                {doc.type === 'employee_badge' && doc.fields.picture && (
                  <div className="badge-picture">
                    <img
                      src={doc.fields.picture}
                      alt="Employee Badge"
                      className="badge-photo"
                    />
                  </div>
                )}
                <div className="document-fields">
                  {Object.entries(doc.fields).map(([key, value]) => {
                    // Don't render picture field as text for employee badge
                    if (key === 'picture' && doc.type === 'employee_badge') return null;

                    // Format label: replace underscores with spaces and add space before numbers
                    const formattedLabel = key
                      .replace(/_/g, ' ')
                      .replace(/(\d+)/g, ' $1');

                    return (
                      <div key={key} className="field-row">
                        <span className="field-label">{formattedLabel}</span>
                        <span className="field-value">{value}</span>
                      </div>
                    );
                  })}
                </div>
              </div>
            ))}
          </div>

          {/* Rules Display */}
          <div className="rules-compact">
            <div className="rules-header">TODAY'S RULES</div>
            <ul className="rules-list">
              {rules.map((rule, i) => (
                <li key={i} className="rule-item">{rule}</li>
              ))}
            </ul>
          </div>
        </div>

        {/* CENTER PANEL - Worker Face & Chat */}
        <div className="center-panel">
          <div className="worker-area">
            {/* Character Portrait */}
            <div className="worker-portrait">
              <img
                src={`https://api.dicebear.com/9.x/notionists-neutral/svg?seed=${currentCase.caseId}&backgroundColor=1a3a52`}
                alt={currentCase.npc.name}
                className="portrait-image"
              />
            </div>

            {/* Worker Name */}
            <div className="worker-name">{currentCase.npc.name}</div>
            <div className="worker-role">{currentCase.npc.role}</div>

            {/* Dialogue Box */}
            <div className="dialogue-box">
              <div className="dialogue-text">{currentCase.openingLine}</div>
              {npcResponse && (
                <div className="dialogue-text dialogue-response">{npcResponse}</div>
              )}
            </div>

            {/* Question Input */}
            <div className="question-area">
              <input
                ref={questionInputRef}
                type="text"
                value={question}
                onChange={(e) => setQuestion(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleAskQuestion()}
                placeholder="Ask a question... (SPACE to focus)"
                disabled={loading || showVerdict}
                className="question-input"
              />
              <button
                onClick={handleAskQuestion}
                disabled={loading || showVerdict || !question.trim()}
                className="ask-button"
              >
                ASK
              </button>
            </div>
          </div>
        </div>

        {/* RIGHT PANEL - Decision Buttons */}
        <div className="right-panel">
          <div className="panel-header">DECISION</div>

          {!showVerdict ? (
            <div className="decision-buttons">
              <button
                className="decision-btn btn-approve"
                onClick={() => handleDecision(Decision.APPROVE)}
                disabled={loading}
              >
                <div className="btn-icon">✓</div>
                <div className="btn-text">APPROVE</div>
              </button>

              <button
                className="decision-btn btn-deny"
                onClick={() => handleDecision(Decision.DENY)}
                disabled={loading}
              >
                <div className="btn-icon">✗</div>
                <div className="btn-text">DENY</div>
              </button>
            </div>
          ) : (
            <div className="verdict-area">
              <div className="verdict-text">{verdict}</div>
              <button onClick={handleNextCase} className="next-case-btn">
                NEXT CASE →
              </button>
            </div>
          )}
        </div>
      </div>
    </>
  );
}

export default GameDesk;
