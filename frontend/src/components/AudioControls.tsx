import { useState, useEffect } from 'react';
import { audioManager } from '../audio/AudioManager';
import './AudioControls.css';

function AudioControls() {
  const [isMuted, setIsMuted] = useState(false);
  const [previousMusicVolume, setPreviousMusicVolume] = useState(0.3);
  const [previousSFXVolume, setPreviousSFXVolume] = useState(0.4);

  useEffect(() => {
    // Check localStorage for saved mute preference
    const savedMuteState = localStorage.getItem('audioMuted');
    if (savedMuteState === 'true') {
      setIsMuted(true);
      audioManager.setMusicVolume(0);
      audioManager.setSFXVolume(0);
    }
  }, []);

  const toggleMute = () => {
    const newMutedState = !isMuted;
    setIsMuted(newMutedState);

    if (newMutedState) {
      // Mute: save current volumes and set to 0
      audioManager.setMusicVolume(0);
      audioManager.setSFXVolume(0);
    } else {
      // Unmute: restore previous volumes
      audioManager.setMusicVolume(previousMusicVolume);
      audioManager.setSFXVolume(previousSFXVolume);
    }

    // Save preference to localStorage
    localStorage.setItem('audioMuted', String(newMutedState));

    // Play button click sound if unmuting
    if (!newMutedState) {
      audioManager.playButtonClick();
    }
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
