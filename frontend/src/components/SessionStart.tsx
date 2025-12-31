import { useState } from 'react';
import { gameClient } from '../api/client';
import { audioManager } from '../audio/AudioManager';
import './SessionStart.css';

interface SessionStartProps {
  onSessionReady: (data: {
    sessionId: string;
    gameDate: string;
    rules: string[];
    totalCases: number;
  }) => void;
}

function SessionStart({ onSessionReady }: SessionStartProps) {
  const [loading, setLoading] = useState(false);
  const [progress, setProgress] = useState({ current: 0, total: 0, message: '' });

  const handleStartSession = async () => {
    setLoading(true);
    setProgress({ current: 0, total: 0, message: 'Initializing...' });

    try {
      const stream: any = gameClient.startSession({ numCases: 2 });

      // Stream session start progress
      for await (const response of stream) {
        if (response.update.case === 'progress' && response.update.value) {
          setProgress({
            current: response.update.value.current,
            total: response.update.value.total,
            message: response.update.value.message,
          });
        } else if (response.update.case === 'ready' && response.update.value) {
          const ready = response.update.value;
          onSessionReady({
            sessionId: ready.sessionId,
            gameDate: ready.gameDate,
            rules: ready.rules,
            totalCases: ready.totalCases,
          });
        }
      }
    } catch (error) {
      console.error('Failed to start session:', error);
      alert('Failed to start session. Check console for details.');
      setLoading(false);
    }
  };

  return (
    <div className="session-start">
      <div className="briefing-card">
        <h2>Transit Desk Briefing</h2>
        <p>
          You are the Transit Clerk at Asteroid Mining Station Delta-7.
          Your job is to verify workers boarding the last shuttle home.
        </p>
        <p>
          Inspect their documents carefully. Approve valid workers, deny invalid ones.
          Make too many mistakes and the company will have your head.
        </p>

        {!loading && (
          <button
            onClick={() => {
              audioManager.playButtonClick();
              handleStartSession();
            }}
            className="start-button"
          >
            Start Shift
          </button>
        )}

        {loading && (
          <div className="loading-screen">
            <div className="progress-bar">
              <div
                className="progress-fill"
                style={{
                  width: progress.total > 0
                    ? `${(progress.current / progress.total) * 100}%`
                    : '0%',
                }}
              />
            </div>
            <p className="progress-text">{progress.message}</p>
            <p className="progress-count">
              {progress.current} / {progress.total}
            </p>
          </div>
        )}
      </div>
    </div>
  );
}

export default SessionStart;
