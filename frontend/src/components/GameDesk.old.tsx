import { useState, useEffect } from 'react';
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

function GameDesk({ sessionId, rules, totalCases, onComplete }: GameDeskProps) {
  const [currentCase, setCurrentCase] = useState<CaseData | null>(null);
  const [score, setScore] = useState(0);
  const [question, setQuestion] = useState('');
  const [npcResponse, setNpcResponse] = useState('');
  const [verdict, setVerdict] = useState('');
  const [showVerdict, setShowVerdict] = useState(false);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadNextCase();
  }, []);

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

      // Play new case loaded sound
      audioManager.playNewCaseSound();
    } catch (error: any) {
      if (error.message?.includes('no more cases')) {
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

    // Play question ask sound
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
        }
        // TODO: Handle audio chunks when ElevenLabs is integrated
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

    // Play appropriate sound for decision
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
    } catch (error) {
      console.error('Failed to resolve case:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleNextCase = () => {
    loadNextCase();
  };

  if (loading && !currentCase) {
    return <div className="loading">Loading next case...</div>;
  }

  if (!currentCase) {
    return null;
  }

  return (
    <div className="game-desk">
      {/* Header with rules and status */}
      <div className="desk-header">
        <div className="rules-panel">
          <h3>Rules of the Day</h3>
          <ul>
            {rules.map((rule, i) => (
              <li key={i}>{rule}</li>
            ))}
          </ul>
        </div>
        <div className="status-panel">
          <div className="status-item">
            <span>Case:</span> {currentCase.caseNumber} / {totalCases}
          </div>
          <div className="status-item">
            <span>Score:</span> {score}
          </div>
          <div className="status-item">
            <span>Secondary Checks:</span> {currentCase.remainingSecondaryChecks}
          </div>
        </div>
      </div>

      {/* NPC Panel */}
      <div className="npc-panel">
        <div className="npc-info">
          <h3>{currentCase.npc.name}</h3>
          <p className="npc-role">{currentCase.npc.role} - {currentCase.npc.department}</p>
          <p className="npc-personality">({currentCase.npc.personality}, {currentCase.npc.demeanor})</p>
        </div>
        <div className="npc-dialogue">
          <p className="opening-line">"{currentCase.openingLine}"</p>
          {npcResponse && (
            <p className="npc-response">"{npcResponse}"</p>
          )}
        </div>
      </div>

      {/* Documents */}
      <div className="documents-area">
        <h3>Documents</h3>
        <div className="documents-grid">
          {currentCase.documents.map((doc, i) => (
            <div key={i} className="document">
              <h4>{doc.type.replace('_', ' ').toUpperCase()}</h4>
              <div className="document-fields">
                {Object.entries(doc.fields).map(([key, value]) => (
                  <div key={key} className="field-row">
                    <span className="field-label">{key}:</span>
                    <span className="field-value">{value}</span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Q&A Section */}
      <div className="qa-section">
        <h3>Ask Question</h3>
        <div className="qa-input">
          <input
            type="text"
            value={question}
            onChange={(e) => setQuestion(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && handleAskQuestion()}
            placeholder="Type your question..."
            disabled={loading || showVerdict}
          />
          <button onClick={handleAskQuestion} disabled={loading || showVerdict || !question.trim()}>
            Ask
          </button>
        </div>
      </div>

      {/* Decision Buttons */}
      {!showVerdict && (
        <div className="decision-buttons">
          <button
            className="btn-approve"
            onClick={() => handleDecision(Decision.APPROVE)}
            disabled={loading}
          >
            ✅ Approve
          </button>
          <button
            className="btn-deny"
            onClick={() => handleDecision(Decision.DENY)}
            disabled={loading}
          >
            ❌ Deny
          </button>
          <button
            className="btn-secondary"
            onClick={() => {
              audioManager.playSecondaryCheckSound();
              alert('Secondary check not yet implemented');
            }}
            disabled={loading || currentCase.remainingSecondaryChecks <= 0}
          >
            🔍 Secondary Check ({currentCase.remainingSecondaryChecks})
          </button>
        </div>
      )}

      {/* Verdict */}
      {showVerdict && (
        <div className="verdict-panel">
          <h3>Verdict</h3>
          <p>{verdict}</p>
          <button
            onClick={() => {
              audioManager.playButtonClick();
              handleNextCase();
            }}
            className="btn-next"
          >
            Next Case →
          </button>
        </div>
      )}
    </div>
  );
}

export default GameDesk;
