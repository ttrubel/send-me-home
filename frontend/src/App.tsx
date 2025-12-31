import { useState, useEffect } from 'react';
import SessionStart from './components/SessionStart';
import GameDesk from './components/GameDesk';
import AudioControls from './components/AudioControls';
import { audioManager } from './audio/AudioManager';
import './App.css';

type GameState = 'start' | 'playing' | 'complete';

interface SessionData {
  sessionId: string;
  gameDate: string;
  rules: string[];
  totalCases: number;
}

interface CompletionStats {
  totalScore: number;
  correct: number;
  incorrect: number;
  accuracy: number;
}

function App() {
  const [gameState, setGameState] = useState<GameState>('start');
  const [sessionData, setSessionData] = useState<SessionData | null>(null);
  const [completionStats, setCompletionStats] = useState<CompletionStats | null>(null);
  const [audioInitialized, setAudioInitialized] = useState(false);

  useEffect(() => {
    // Initialize audio on first user interaction
    const initAudio = async () => {
      if (!audioInitialized) {
        await audioManager.initialize();
        // Only play background music if not muted
        if (!audioManager.isMuted()) {
          audioManager.playBackgroundMusic();
        }
        setAudioInitialized(true);
      }
    };

    // Start audio on any user interaction
    const handleInteraction = () => {
      initAudio();
      // Remove listeners after first interaction
      document.removeEventListener('click', handleInteraction);
      document.removeEventListener('keydown', handleInteraction);
    };

    document.addEventListener('click', handleInteraction);
    document.addEventListener('keydown', handleInteraction);

    return () => {
      document.removeEventListener('click', handleInteraction);
      document.removeEventListener('keydown', handleInteraction);
    };
  }, [audioInitialized]);

  const handleSessionReady = (data: SessionData) => {
    setSessionData(data);
    setGameState('playing');
  };

  const handleSessionComplete = (stats: CompletionStats) => {
    setCompletionStats(stats);
    setGameState('complete');
  };

  return (
    <div className="app">
      <AudioControls />

      <header className="app-header">
        <h1>⛏️ SEND ME HOME</h1>
        <p className="subtitle">Transit Desk - Asteroid Mining Station Delta-7</p>
      </header>

      <main className="app-main">
        {gameState === 'start' && (
          <SessionStart onSessionReady={handleSessionReady} />
        )}

        {gameState === 'playing' && sessionData && (
          <GameDesk
            sessionId={sessionData.sessionId}
            gameDate={sessionData.gameDate}
            rules={sessionData.rules}
            totalCases={sessionData.totalCases}
            onComplete={handleSessionComplete}
          />
        )}

        {gameState === 'complete' && completionStats && (
          <div className="session-complete">
            <h2>🎯 Shift Complete</h2>
            <div className="final-stats">
              <div className="stat-row">
                <span className="stat-label">Final Score:</span>
                <span className="stat-value">{completionStats.totalScore}</span>
              </div>
              <div className="stat-row">
                <span className="stat-label">Accuracy:</span>
                <span className="stat-value">{completionStats.accuracy}%</span>
              </div>
              <div className="stat-row">
                <span className="stat-label">Correct Decisions:</span>
                <span className="stat-value">{completionStats.correct}</span>
              </div>
              <div className="stat-row">
                <span className="stat-label">Incorrect Decisions:</span>
                <span className="stat-value">{completionStats.incorrect}</span>
              </div>
            </div>
            <p className="completion-message">
              {completionStats.accuracy >= 80
                ? "Excellent work, clerk! The station runs smoothly with you on duty."
                : completionStats.accuracy >= 60
                  ? "Decent work. You'll get better with practice."
                  : "That was rough. Maybe review the rules next time?"}
            </p>
            <button
              onClick={() => {
                audioManager.playButtonClick();
                setGameState('start');
                setCompletionStats(null);
              }}
            >
              New Shift
            </button>
          </div>
        )}
      </main>
    </div>
  );
}

export default App;
