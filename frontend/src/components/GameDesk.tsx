import { useState, useEffect, useRef } from 'react';
import { gameClient } from '../api/client';
import { Decision } from '../gen/game/v1/game_pb';
import type { NPCProfile, Document } from '../gen/game/v1/game_pb';
import { audioManager } from '../audio/AudioManager';
import './GameDesk.css';

interface GameDeskProps {
  sessionId: string;
  rules: string[];
  totalCases: number;
  onComplete: () => void;
}

interface CaseData {
  caseId: string;
  npc: NPCProfile;
  documents: Document[];
  openingLine: string;
  caseNumber: number;
  remainingSecondaryChecks: number;
}

interface LogEntry {
  id: number;
  time: string;
  action: string;
  type: 'approved' | 'denied' | 'info';
}

function GameDesk({ sessionId, rules, totalCases, onComplete }: GameDeskProps) {
  const [currentCase, setCurrentCase] = useState<CaseData | null>(null);
  const [score, setScore] = useState(0);
  const [question, setQuestion] = useState('');
  const [npcResponse, setNpcResponse] = useState('');
  const [verdict, setVerdict] = useState('');
  const [showVerdict, setShowVerdict] = useState(false);
  const [loading, setLoading] = useState(false);
  const [shiftLog, setShiftLog] = useState<LogEntry[]>([]);
  const logIdCounter = useRef(0);
  const questionInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    loadNextCase();
    addLogEntry('Shift started. Transit desk operational.', 'info');
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

  const addLogEntry = (action: string, type: 'approved' | 'denied' | 'info' = 'info') => {
    const now = new Date();
    const time = now.toLocaleTimeString('en-US', { hour12: false });
    logIdCounter.current += 1;
    setShiftLog(prev => [
      { id: logIdCounter.current, time, action, type },
      ...prev.slice(0, 19) // Keep last 20 entries
    ]);
  };

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
      addLogEntry(`Case #${response.caseNumber}: ${response.npc!.name} - ${response.npc!.role}`, 'info');
    } catch (error: any) {
      if (error.message?.includes('no more cases')) {
        addLogEntry('All cases processed. Shift complete.', 'info');
        onComplete();
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
    addLogEntry(`Q: ${question}`, 'info');

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

      setVerdict(response.verdict);
      setScore(response.totalScore);
      setShowVerdict(true);

      const decisionText = decision === Decision.APPROVE ? 'APPROVED' : 'DENIED';
      const logType = decision === Decision.APPROVE ? 'approved' : 'denied';
      addLogEntry(`${currentCase.npc.name}: ${decisionText}`, logType);
    } catch (error) {
      console.error('Failed to resolve case:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleNextCase = () => {
    audioManager.playButtonClick();
    loadNextCase();
  };

  // Calculate shuttle capacity (visual indicator)
  const shuttleCapacity = 12;
  const occupiedSeats = Math.min(currentCase?.caseNumber || 0, shuttleCapacity);

  if (loading && !currentCase) {
    return <div className="loading">▮▮▮ LOADING NEXT CASE ▮▮▮</div>;
  }

  if (!currentCase) {
    return null;
  }

  return (
    <>
      <div className="game-desk">
        {/* TOP BAR - Shuttle Status */}
        <div className="shuttle-status-bar">
          <div className="shuttle-display">
            <span className="shuttle-label">◈ SHUTTLE STATUS</span>
            <div className="shuttle-leds">
              {Array.from({ length: shuttleCapacity }).map((_, i) => (
                <div
                  key={i}
                  className={`shuttle-led ${i < occupiedSeats ? 'active' : ''}`}
                  title={`Seat ${i + 1}`}
                />
              ))}
            </div>
          </div>
          <div className="queue-display">
            <span>QUEUE:</span>
            <span className="queue-count">{totalCases - currentCase.caseNumber + 1}</span>
            <span>REMAINING</span>
          </div>
        </div>

        {/* LEFT PANEL - Clerk's Tools */}
        <div className="clerk-panel">
          {/* Clearance Badge */}
          <div className="clearance-badge">
            <div className="badge-header">
              <div className="badge-title">◈ TRANSIT AUTHORITY ◈</div>
              <div className="badge-station">DELTA-7 STATION</div>
            </div>
            <div className="badge-info">
              <div className="badge-row">
                <span className="badge-label">CLERK ID:</span>
                <span className="badge-value">TC-{sessionId.slice(0, 8).toUpperCase()}</span>
              </div>
              <div className="badge-row">
                <span className="badge-label">CLEARANCE:</span>
                <span className="badge-value">LEVEL 3</span>
              </div>
              <div className="badge-row">
                <span className="badge-label">SHIFT:</span>
                <span className="badge-value">FINAL DEPARTURE</span>
              </div>
            </div>
          </div>

          {/* Rules Book */}
          <div className="rules-book">
            <div className="rules-book-header">
              RULES OF THE DAY
            </div>
            <ul className="rules-list">
              {rules.map((rule, i) => (
                <li key={i} className="rule-item">{rule}</li>
              ))}
            </ul>
          </div>
        </div>

        {/* CENTER PANEL - NPC Window & Documents */}
        <div className="center-panel">
          {/* NPC Window */}
          <div className="npc-window">
            <div className="npc-screen-header">
              <div className="npc-name">{currentCase.npc.name}</div>
              <div className="npc-id">ID: {currentCase.caseId.slice(0, 8)}</div>
            </div>

            <div className="npc-content">
              {/* Character Portrait */}
              <div className="npc-portrait">
                <img
                  src={`https://api.dicebear.com/7.x/bottts/svg?seed=${currentCase.caseId}&backgroundColor=1a3a52&scale=90`}
                  alt={currentCase.npc.name}
                  className="portrait-image"
                />
                <div className="portrait-frame"></div>
              </div>

              <div className="npc-details">
              <div className="npc-detail-item">
                <div className="npc-detail-label">Position</div>
                <div className="npc-detail-value">{currentCase.npc.role}</div>
              </div>
              <div className="npc-detail-item">
                <div className="npc-detail-label">Department</div>
                <div className="npc-detail-value">{currentCase.npc.department}</div>
              </div>
              <div className="npc-detail-item">
                <div className="npc-detail-label">Personality</div>
                <div className="npc-detail-value">{currentCase.npc.personality}</div>
              </div>
              <div className="npc-detail-item">
                <div className="npc-detail-label">Demeanor</div>
                <div className="npc-detail-value">{currentCase.npc.demeanor}</div>
              </div>
              </div>
            </div>

            <div className="npc-dialogue-box">
              <div className="dialogue-text">{currentCase.openingLine}</div>
              {npcResponse && (
                <div className="dialogue-text dialogue-response">{npcResponse}</div>
              )}
            </div>
          </div>

          {/* Documents Area */}
          <div className="documents-area">
            <div className="documents-header">
              CONTRACT DOCUMENTS
            </div>
            <div className="documents-grid">
              {currentCase.documents.map((doc, i) => (
                <div key={i} className="document">
                  <div className="document-type">
                    {doc.type.replace('_', ' ')}
                  </div>
                  <div className="document-fields">
                    {Object.entries(doc.fields).map(([key, value]) => (
                      <div key={key} className="field-row">
                        <span className="field-label">{key}</span>
                        <span className="field-value">{value}</span>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* RIGHT PANEL - Shift Log & Contract Details */}
        <div className="right-panel">
          {/* Contract Details / Stats */}
          <div className="contract-details">
            <div className="contract-header">◈ SHIFT STATISTICS ◈</div>
            <div className="contract-stats">
              <div className="stat-box">
                <div className="stat-label">Case</div>
                <div className="stat-value">{currentCase.caseNumber}/{totalCases}</div>
              </div>
              <div className="stat-box">
                <div className="stat-label">Score</div>
                <div className="stat-value score">{score}</div>
              </div>
              <div className="stat-box">
                <div className="stat-label">Checks</div>
                <div className="stat-value checks">{currentCase.remainingSecondaryChecks}</div>
              </div>
              <div className="stat-box">
                <div className="stat-label">Queue</div>
                <div className="stat-value">{totalCases - currentCase.caseNumber + 1}</div>
              </div>
            </div>
          </div>

          {/* Shift Log */}
          <div className="shift-log">
            <div className="shift-log-header">
              SHIFT LOG
            </div>
            <div className="log-entries">
              {shiftLog.map(entry => (
                <div key={entry.id} className={`log-entry ${entry.type}`}>
                  <div className="log-time">[{entry.time}]</div>
                  <div className="log-action">{entry.action}</div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* BOTTOM PANEL - Actions */}
        <div className="action-panel">
          {/* Question Input */}
          <div className="qa-section">
            <div className="qa-label">
              ◈ INTERROGATE WORKER
              <span className="space-indicator-inline">
                <kbd>SPACE</kbd> TO SPEAK
              </span>
            </div>
            <div className="qa-input-wrapper">
              <input
                ref={questionInputRef}
                type="text"
                value={question}
                onChange={(e) => setQuestion(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleAskQuestion()}
                placeholder="Type your question..."
                disabled={loading || showVerdict}
              />
              <button
                onClick={handleAskQuestion}
                disabled={loading || showVerdict || !question.trim()}
                className="action-button"
                style={{ width: '80px', fontSize: '14px' }}
              >
                ASK
              </button>
            </div>
          </div>

          {/* Action Buttons */}
          {!showVerdict && (
            <>
              <button
                className="action-button btn-approve"
                onClick={() => handleDecision(Decision.APPROVE)}
                disabled={loading}
              >
                <div className="button-icon">✓</div>
                <div className="button-label">Approve</div>
              </button>

              <button
                className="action-button btn-deny"
                onClick={() => handleDecision(Decision.DENY)}
                disabled={loading}
              >
                <div className="button-icon">✗</div>
                <div className="button-label">Deny</div>
              </button>

              <button
                className="action-button btn-secondary"
                onClick={() => {
                  audioManager.playSecondaryCheckSound();
                  addLogEntry('Secondary biometric scan initiated', 'info');
                  alert('Secondary check not yet implemented');
                }}
                disabled={loading || currentCase.remainingSecondaryChecks <= 0}
              >
                <div className="button-icon">🔍</div>
                <div className="button-label">Secondary</div>
                <div className="button-quota">({currentCase.remainingSecondaryChecks})</div>
              </button>
            </>
          )}
        </div>
      </div>

      {/* Verdict Overlay */}
      {showVerdict && (
        <div className="verdict-overlay">
          <div className="verdict-panel">
            <div className="verdict-header">◈ VERDICT ◈</div>
            <div className="verdict-text">{verdict}</div>
            <button onClick={handleNextCase} className="btn-next">
              NEXT CASE →
            </button>
          </div>
        </div>
      )}
    </>
  );
}

export default GameDesk;
