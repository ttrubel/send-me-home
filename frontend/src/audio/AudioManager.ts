/**
 * AudioManager - Handles retro background music and UI sound effects
 * Uses Web Audio API to generate synth sounds fitting the sci-fi theme
 */

export class AudioManager {
  private audioContext: AudioContext | null = null;
  private backgroundMusic: AudioBufferSourceNode | null = null;
  private musicGain: GainNode | null = null;
  private sfxGain: GainNode | null = null;
  private isInitialized = false;
  private isMusicPlaying = false;

  constructor() {
    // Audio context must be created after user interaction
  }

  /**
   * Initialize audio context (call after user interaction)
   */
  async initialize() {
    if (this.isInitialized) return;

    this.audioContext = new AudioContext();

    // Create gain nodes for volume control
    this.musicGain = this.audioContext.createGain();
    this.musicGain.gain.value = 0.3; // Background music at 30%
    this.musicGain.connect(this.audioContext.destination);

    this.sfxGain = this.audioContext.createGain();
    this.sfxGain.gain.value = 0.4; // Sound effects at 40%
    this.sfxGain.connect(this.audioContext.destination);

    this.isInitialized = true;
  }

  /**
   * Generate retro synthwave background music loop
   */
  private generateBackgroundMusic(): AudioBuffer {
    if (!this.audioContext) throw new Error('Audio context not initialized');

    const sampleRate = this.audioContext.sampleRate;
    const duration = 8; // 8 second loop
    const buffer = this.audioContext.createBuffer(2, sampleRate * duration, sampleRate);

    // Generate a retro sci-fi atmosphere track
    for (let channel = 0; channel < 2; channel++) {
      const data = buffer.getChannelData(channel);

      // Base drone (low synth pad)
      const baseFreq = 55; // A1 note
      for (let i = 0; i < data.length; i++) {
        const t = i / sampleRate;

        // Bass drone with subtle movement
        const bass = Math.sin(2 * Math.PI * baseFreq * t) * 0.15;
        const bassHarmonic = Math.sin(2 * Math.PI * baseFreq * 1.5 * t) * 0.08;

        // Arpeggio pattern (classic sci-fi computer feel)
        const arpPattern = [0, 4, 7, 12, 7, 4]; // A minor arpeggio
        const arpIndex = Math.floor((t * 2) % arpPattern.length);
        const arpNote = baseFreq * Math.pow(2, arpPattern[arpIndex] / 12);
        const arp = Math.sin(2 * Math.PI * arpNote * t) * 0.1;

        // Add envelope to arp notes
        const arpEnv = Math.max(0, 1 - ((t * 2) % 0.5) * 4);

        // Ambient high pad
        const pad = Math.sin(2 * Math.PI * baseFreq * 4 * t) * 0.05;

        // Subtle LFO for movement
        const lfo = Math.sin(2 * Math.PI * 0.25 * t);

        data[i] = bass + bassHarmonic + (arp * arpEnv) + (pad * (0.5 + lfo * 0.5));
      }
    }

    return buffer;
  }

  /**
   * Start playing background music
   */
  playBackgroundMusic() {
    if (!this.isInitialized || this.isMusicPlaying || !this.audioContext || !this.musicGain) {
      return;
    }

    try {
      const buffer = this.generateBackgroundMusic();
      this.backgroundMusic = this.audioContext.createBufferSource();
      this.backgroundMusic.buffer = buffer;
      this.backgroundMusic.loop = true;
      this.backgroundMusic.connect(this.musicGain);
      this.backgroundMusic.start(0);
      this.isMusicPlaying = true;
    } catch (error) {
      console.error('Failed to start background music:', error);
    }
  }

  /**
   * Stop background music
   */
  stopBackgroundMusic() {
    if (this.backgroundMusic) {
      this.backgroundMusic.stop();
      this.backgroundMusic.disconnect();
      this.backgroundMusic = null;
      this.isMusicPlaying = false;
    }
  }

  /**
   * Play button click sound (satisfying retro beep)
   */
  playButtonClick() {
    if (!this.isInitialized || !this.audioContext || !this.sfxGain) return;

    const now = this.audioContext.currentTime;
    const oscillator = this.audioContext.createOscillator();
    const gainNode = this.audioContext.createGain();

    oscillator.connect(gainNode);
    gainNode.connect(this.sfxGain);

    // Retro computer beep - descending pitch
    oscillator.frequency.setValueAtTime(800, now);
    oscillator.frequency.exponentialRampToValueAtTime(600, now + 0.05);

    oscillator.type = 'square'; // Retro square wave

    // Quick envelope
    gainNode.gain.setValueAtTime(0.3, now);
    gainNode.gain.exponentialRampToValueAtTime(0.01, now + 0.1);

    oscillator.start(now);
    oscillator.stop(now + 0.1);
  }

  /**
   * Play approve button sound (positive ascending beep)
   */
  playApproveSound() {
    if (!this.isInitialized || !this.audioContext || !this.sfxGain) return;

    const now = this.audioContext.currentTime;
    const oscillator = this.audioContext.createOscillator();
    const gainNode = this.audioContext.createGain();

    oscillator.connect(gainNode);
    gainNode.connect(this.sfxGain);

    // Positive ascending tones
    oscillator.frequency.setValueAtTime(600, now);
    oscillator.frequency.exponentialRampToValueAtTime(900, now + 0.1);

    oscillator.type = 'sine';

    gainNode.gain.setValueAtTime(0.4, now);
    gainNode.gain.exponentialRampToValueAtTime(0.01, now + 0.15);

    oscillator.start(now);
    oscillator.stop(now + 0.15);
  }

  /**
   * Play deny button sound (negative descending buzz)
   */
  playDenySound() {
    if (!this.isInitialized || !this.audioContext || !this.sfxGain) return;

    const now = this.audioContext.currentTime;
    const oscillator = this.audioContext.createOscillator();
    const gainNode = this.audioContext.createGain();

    oscillator.connect(gainNode);
    gainNode.connect(this.sfxGain);

    // Negative descending buzz
    oscillator.frequency.setValueAtTime(400, now);
    oscillator.frequency.exponentialRampToValueAtTime(200, now + 0.15);

    oscillator.type = 'sawtooth'; // Harsher tone

    gainNode.gain.setValueAtTime(0.4, now);
    gainNode.gain.exponentialRampToValueAtTime(0.01, now + 0.2);

    oscillator.start(now);
    oscillator.stop(now + 0.2);
  }

  /**
   * Play secondary check sound (scan effect)
   */
  playSecondaryCheckSound() {
    if (!this.isInitialized || !this.audioContext || !this.sfxGain) return;

    const now = this.audioContext.currentTime;
    const oscillator = this.audioContext.createOscillator();
    const gainNode = this.audioContext.createGain();

    oscillator.connect(gainNode);
    gainNode.connect(this.sfxGain);

    // Scanning sweep
    oscillator.frequency.setValueAtTime(300, now);
    oscillator.frequency.exponentialRampToValueAtTime(1200, now + 0.25);
    oscillator.frequency.exponentialRampToValueAtTime(300, now + 0.5);

    oscillator.type = 'triangle';

    gainNode.gain.setValueAtTime(0.3, now);
    gainNode.gain.setValueAtTime(0.3, now + 0.4);
    gainNode.gain.exponentialRampToValueAtTime(0.01, now + 0.5);

    oscillator.start(now);
    oscillator.stop(now + 0.5);
  }

  /**
   * Play new case loaded sound (data transfer)
   */
  playNewCaseSound() {
    if (!this.isInitialized || !this.audioContext || !this.sfxGain) return;

    const now = this.audioContext.currentTime;

    // Play a quick sequence of blips
    for (let i = 0; i < 3; i++) {
      const oscillator = this.audioContext.createOscillator();
      const gainNode = this.audioContext.createGain();

      oscillator.connect(gainNode);
      gainNode.connect(this.sfxGain);

      const startTime = now + (i * 0.08);
      oscillator.frequency.setValueAtTime(700 + (i * 200), startTime);
      oscillator.type = 'square';

      gainNode.gain.setValueAtTime(0.2, startTime);
      gainNode.gain.exponentialRampToValueAtTime(0.01, startTime + 0.06);

      oscillator.start(startTime);
      oscillator.stop(startTime + 0.06);
    }
  }

  /**
   * Play question ask sound (communication beep)
   */
  playAskQuestionSound() {
    if (!this.isInitialized || !this.audioContext || !this.sfxGain) return;

    const now = this.audioContext.currentTime;
    const oscillator = this.audioContext.createOscillator();
    const gainNode = this.audioContext.createGain();

    oscillator.connect(gainNode);
    gainNode.connect(this.sfxGain);

    oscillator.frequency.setValueAtTime(1000, now);
    oscillator.type = 'sine';

    gainNode.gain.setValueAtTime(0.3, now);
    gainNode.gain.exponentialRampToValueAtTime(0.01, now + 0.08);

    oscillator.start(now);
    oscillator.stop(now + 0.08);
  }

  /**
   * Set music volume (0-1)
   */
  setMusicVolume(volume: number) {
    if (this.musicGain) {
      this.musicGain.gain.value = Math.max(0, Math.min(1, volume));
    }
  }

  /**
   * Set sound effects volume (0-1)
   */
  setSFXVolume(volume: number) {
    if (this.sfxGain) {
      this.sfxGain.gain.value = Math.max(0, Math.min(1, volume));
    }
  }

  /**
   * Cleanup
   */
  dispose() {
    this.stopBackgroundMusic();
    if (this.audioContext) {
      this.audioContext.close();
    }
  }
}

// Singleton instance
export const audioManager = new AudioManager();
