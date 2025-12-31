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

function App() {
  const [gameState, setGameState] = useState<GameState>('start');
  const [sessionData, setSessionData] = useState<SessionData | null>(null);
  const [audioInitialized, setAudioInitialized] = useState(false);

  useEffect(() => {
    // Initialize audio on first user interaction
    const initAudio = async () => {
      if (!audioInitialized) {
        await audioManager.initialize();
        audioManager.playBackgroundMusic();
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

  const handleSessionComplete = () => {
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

        {gameState === 'complete' && (
          <div className="session-complete">
            <h2>Shift Complete</h2>
            <p>All cases processed. Good work, clerk.</p>
            <button
              onClick={() => {
                audioManager.playButtonClick();
                setGameState('start');
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
