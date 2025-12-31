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
  private defaultMusicVolume = 0.3;
  private defaultSFXVolume = 0.4;

  constructor() {
    // Audio context must be created after user interaction
  }

  /**
   * Check if audio is muted from localStorage
   */
  isMuted(): boolean {
    return localStorage.getItem('audioMuted') === 'true';
  }

  /**
   * Get the default music volume (used when unmuting)
   */
  getDefaultMusicVolume(): number {
    return this.defaultMusicVolume;
  }

  /**
   * Get the default SFX volume (used when unmuting)
   */
  getDefaultSFXVolume(): number {
    return this.defaultSFXVolume;
  }

  /**
   * Initialize audio context (call after user interaction)
   */
  async initialize() {
    if (this.isInitialized) return;

    this.audioContext = new AudioContext();

    // Create gain nodes for volume control
    this.musicGain = this.audioContext.createGain();
    this.sfxGain = this.audioContext.createGain();

    // Set initial volumes based on mute state
    const muted = this.isMuted();
    this.musicGain.gain.value = muted ? 0 : this.defaultMusicVolume;
    this.sfxGain.gain.value = muted ? 0 : this.defaultSFXVolume;

    this.musicGain.connect(this.audioContext.destination);
    this.sfxGain.connect(this.audioContext.destination);

    this.isInitialized = true;
  }

  /**
   * Generate square wave (classic 8-bit sound)
   */
  private squareWave(t: number, freq: number): number {
    return Math.sign(Math.sin(2 * Math.PI * freq * t));
  }

  /**
   * Generate triangle wave (softer 8-bit sound)
   */
  private triangleWave(t: number, freq: number): number {
    return 2 * Math.abs(2 * ((t * freq) - Math.floor((t * freq) + 0.5))) - 1;
  }

  /**
   * Generate 8-bit chiptune background music loop
   */
  private generateBackgroundMusic(): AudioBuffer {
    if (!this.audioContext) throw new Error('Audio context not initialized');

    const sampleRate = this.audioContext.sampleRate;
    const duration = 16; // 16 second loop for more complex melody
    const buffer = this.audioContext.createBuffer(2, sampleRate * duration, sampleRate);

    // Note frequencies (A minor scale, great for sci-fi)
    const A3 = 220.00;
    const C4 = 261.63;
    const D4 = 293.66;
    const E4 = 329.63;
    const F4 = 349.23;
    const G4 = 392.00;
    const A4 = 440.00;
    const C5 = 523.25;
    const D5 = 587.33;
    const E5 = 659.25;

    // 8-bit chiptune melody (16 bars, each 1 second)
    // Pattern: intro, main theme, variation, repeat
    const melody: (number | null)[] = [
      A4, null, E4, null, A4, null, C5, null,  // Bar 1-2: Opening phrase
      D5, D5, null, C5, A4, null, null, null,  // Bar 3-4: Descending
      E4, null, G4, null, A4, null, E4, null,  // Bar 5-6: Response phrase
      F4, F4, null, E4, D4, null, null, null,  // Bar 7-8: Resolution
      A4, null, E4, null, A4, null, C5, null,  // Bar 9-10: Repeat theme
      D5, D5, null, E5, D5, C5, null, null,   // Bar 11-12: Higher variation
      A4, null, C5, null, E5, null, D5, null,  // Bar 13-14: Climax
      C5, C5, null, A4, E4, null, null, null   // Bar 15-16: End phrase
    ];

    // Bass line (octave lower, quarter notes)
    const bassLine: (number | null)[] = [
      A3, null, null, null, A3, null, null, null,  // Bar 1-2
      D4, null, null, null, D4, null, null, null,  // Bar 3-4
      E4, null, null, null, E4, null, null, null,  // Bar 5-6
      F4, null, null, null, D4, null, null, null,  // Bar 7-8
      A3, null, null, null, A3, null, null, null,  // Bar 9-10
      D4, null, null, null, D4, null, null, null,  // Bar 11-12
      A3, null, null, null, A3, null, null, null,  // Bar 13-14
      C4, null, null, null, E4, null, null, null   // Bar 15-16
    ];

    // Arpeggio pattern (fast 16th note feel)
    const arpPattern = [0, 4, 7, 12]; // A minor chord intervals

    const beatsPerSecond = 2; // 120 BPM (2 beats per second)
    const noteLength = 0.5; // 8th notes

    for (let channel = 0; channel < 2; channel++) {
      const data = buffer.getChannelData(channel);

      for (let i = 0; i < data.length; i++) {
        const t = i / sampleRate;
        let sample = 0;

        // Lead melody (square wave - classic NES/GameBoy sound)
        const melodyIndex = Math.floor(t / noteLength) % melody.length;
        const melodyNote = melody[melodyIndex];
        if (melodyNote) {
          // ADSR envelope for notes
          const noteTime = t % noteLength;
          const attack = 0.02;
          const decay = 0.05;
          const sustain = 0.6;
          const release = 0.1;

          let envelope = 1;
          if (noteTime < attack) {
            envelope = noteTime / attack;
          } else if (noteTime < attack + decay) {
            envelope = 1 - ((noteTime - attack) / decay) * (1 - sustain);
          } else if (noteTime > noteLength - release) {
            envelope = sustain * ((noteLength - noteTime) / release);
          } else {
            envelope = sustain;
          }

          sample += this.squareWave(t, melodyNote) * envelope * 0.15;
        }

        // Bass line (triangle wave - warmer, deeper)
        const bassIndex = Math.floor(t / noteLength) % bassLine.length;
        const bassNote = bassLine[bassIndex];
        if (bassNote) {
          const noteTime = t % noteLength;
          const envelope = noteTime < 0.05 ? noteTime / 0.05 : 0.8;
          sample += this.triangleWave(t, bassNote) * envelope * 0.2;
        }

        // Arpeggio backing (higher pitch, pulse wave)
        const arpSpeed = 8; // Fast arpeggiation
        const arpIndex = Math.floor(t * arpSpeed) % arpPattern.length;
        const baseNote = bassLine[bassIndex] || A3;
        const arpNote = baseNote * Math.pow(2, arpPattern[arpIndex] / 12) * 2; // Octave up
        const arpEnv = 0.5 + 0.3 * Math.sin(t * 2 * Math.PI * 0.5); // Slow LFO
        sample += this.squareWave(t, arpNote) * 0.08 * arpEnv;

        // Stereo panning for arpeggio (channel 0 = left, channel 1 = right)
        if (channel === 0) {
          sample += this.squareWave(t, arpNote * 1.01) * 0.04 * arpEnv;
        } else {
          sample += this.squareWave(t, arpNote * 0.99) * 0.04 * arpEnv;
        }

        // Subtle noise percussion on beats (for rhythm)
        const beatTime = t * beatsPerSecond;
        const beatFrac = beatTime - Math.floor(beatTime);
        if (beatFrac < 0.05) {
          const noise = (Math.random() * 2 - 1) * beatFrac * 0.1;
          sample += noise;
        }

        data[i] = sample;
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
