# Audio System

This directory contains the audio management system for Send Me Home, providing retro sci-fi background music and UI sound effects.

## Overview

The audio system uses the Web Audio API to generate synthesized retro-style sounds that match the game's sci-fi mining station theme. All audio is procedurally generated - no audio files needed!

## AudioManager

The `AudioManager` class is a singleton that handles all game audio:

```typescript
import { audioManager } from '../audio/AudioManager';
```

### Initialization

Audio must be initialized after user interaction (browser requirement):

```typescript
await audioManager.initialize();
audioManager.playBackgroundMusic();
```

This is handled automatically in `App.tsx` on the first user click or keypress.

## Background Music

The background music is a 16-second looping 8-bit chiptune track in the style of classic NES/Game Boy games, featuring:
- **Lead Melody**: Square wave melody in A minor (classic 8-bit sound)
- **Bass Line**: Triangle wave bass (warmer, deeper tone)
- **Arpeggio Backing**: Fast arpeggiated chords with stereo panning
- **Rhythm**: Subtle noise percussion on beats
- **ADSR Envelopes**: Proper attack/decay/sustain/release for authentic chiptune feel

The melody follows a structured composition:
- Bars 1-4: Opening phrase with descending motif
- Bars 5-8: Response phrase with resolution
- Bars 9-12: Theme variation with higher notes
- Bars 13-16: Climax and ending phrase

Tempo: 120 BPM | Key: A minor | Style: Retro 8-bit chiptune

The music loops seamlessly and plays at 30% volume by default.

## Sound Effects

### Button Click
```typescript
audioManager.playButtonClick();
```
General button press sound - descending square wave beep (800Hz → 600Hz)

### Approve Decision
```typescript
audioManager.playApproveSound();
```
Positive ascending tone for approval (600Hz → 900Hz)

### Deny Decision
```typescript
audioManager.playDenySound();
```
Negative descending buzz for denial (400Hz → 200Hz, sawtooth wave)

### Secondary Check
```typescript
audioManager.playSecondaryCheckSound();
```
Scanning sweep effect (300Hz → 1200Hz → 300Hz)

### New Case Loaded
```typescript
audioManager.playNewCaseSound();
```
Three-note data transfer blip sequence

### Ask Question
```typescript
audioManager.playAskQuestionSound();
```
Communication beep (1000Hz sine wave)

## Audio Controls

### Mute Button

A floating mute button (🔊/🔇) is available in the top-right corner of the screen:
- Click to toggle mute/unmute
- Mute state persists across page reloads (saved in localStorage)
- When unmuted, plays a confirmation sound
- Visual feedback with red tint when muted

### Volume Control (Programmatic)

Adjust volumes independently:

```typescript
audioManager.setMusicVolume(0.5);  // 0-1 range
audioManager.setSFXVolume(0.7);    // 0-1 range
```

## Implementation Notes

### Why Synthesized Audio?

1. **No assets required** - Everything generated in code
2. **Authentic 8-bit aesthetic** - Real square/triangle waves like NES/Game Boy
3. **Tiny bundle size** - No audio files to download (~15KB of code)
4. **Customizable** - Easy to tweak melodies and parameters
5. **True chiptune** - Not samples, but actual waveform synthesis

### Browser Compatibility

The Web Audio API is supported in all modern browsers. Audio context initialization requires user interaction (autoplay policy).

### Performance

All sound effects are generated on-demand and are very lightweight. The 16-second chiptune music buffer is generated once (takes ~50ms) and then looped efficiently with zero CPU overhead during playback.

## Future Enhancements

- ✅ ~~Add mute toggle~~ (Completed!)
- Add volume sliders in UI for fine control
- Generate additional music variations
- Add ambient sound effects (station machinery, etc.)
- Integrate with ElevenLabs voice responses
- Add music track selection menu

## Customization

### Modifying the 8-bit Music

The chiptune melody is defined as arrays in `generateBackgroundMusic()`:

```typescript
// Edit the melody array - each element is a note frequency or null for rest
const melody: (number | null)[] = [
  A4, null, E4, null, A4, null, C5, null,  // Your pattern here
  // ... 32 elements total (16 seconds at 8th notes)
];

// Edit the bass line
const bassLine: (number | null)[] = [
  A3, null, null, null, A3, null, null, null,
  // ... your bass pattern
];
```

### Modifying Sound Effects

Edit frequency, duration, and envelope parameters in `AudioManager.ts`:

```typescript
oscillator.frequency.setValueAtTime(startFreq, now);
oscillator.frequency.exponentialRampToValueAtTime(endFreq, now + duration);
oscillator.type = 'square' | 'sine' | 'sawtooth' | 'triangle';
```

### Wave Types
- **Square**: Harsh, classic 8-bit (NES pulse channels)
- **Triangle**: Warm, bass sounds (NES triangle channel)
- **Sawtooth**: Buzzy, aggressive
- **Sine**: Pure, smooth tones
