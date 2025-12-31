import { useState, useEffect } from 'react';
import { audioManager } from '../audio/AudioManager';
import './AudioControls.css';

function AudioControls() {
  const [isMuted, setIsMuted] = useState(false);

  useEffect(() => {
    // Check localStorage for saved mute preference
    const savedMuteState = localStorage.getItem('audioMuted');
    if (savedMuteState === 'true') {
      setIsMuted(true);
    }
  }, []);

  const toggleMute = () => {
    const newMutedState = !isMuted;
    setIsMuted(newMutedState);

    if (newMutedState) {
      // Mute: stop music and set volumes to 0
      audioManager.stopBackgroundMusic();
      audioManager.setMusicVolume(0);
      audioManager.setSFXVolume(0);
    } else {
      // Unmute: restore volumes and start music
      audioManager.setMusicVolume(audioManager.getDefaultMusicVolume());
      audioManager.setSFXVolume(audioManager.getDefaultSFXVolume());
      audioManager.playBackgroundMusic();
      // Play button click sound
      audioManager.playButtonClick();
    }

    // Save preference to localStorage
    localStorage.setItem('audioMuted', String(newMutedState));
  };

  return (
    <button
      className={`audio-control-btn ${isMuted ? 'muted' : ''}`}
      onClick={toggleMute}
      title={isMuted ? 'Unmute audio' : 'Mute audio'}
      aria-label={isMuted ? 'Unmute audio' : 'Mute audio'}
    >
      {isMuted ? '🔇' : '🔊'}
    </button>
  );
}

export default AudioControls;
