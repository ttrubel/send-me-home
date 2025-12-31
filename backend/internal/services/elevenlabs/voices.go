package elevenlabs

import (
	"hash/fnv"
)

// ElevenLabs Voice IDs for different character types
// These are real voice IDs from ElevenLabs
const (
	// Male voices - varied personalities
	VoiceMaleRough     = "pNInz6obpgDQGcFmaJgB" // Adam - deep, authoritative
	VoiceMaleYoung     = "yoZ06aMxZJJ28mfd3POQ" // Sam - young male
	VoiceMaleCalm      = "TxGEqnHWrfWFTfGW9XjX" // Josh - calm, professional
	VoiceMaleOld       = "VR6AewLTigWG4xSOukaG" // Arnold - older male
	VoiceMaleGruff     = "ErXwobaYiN019PkySvjV" // Antoni - gruff
	VoiceMaleDeep      = "GBv7mTt0atIp3Br8iCZE" // Thomas - deep voice

	// Female voices - varied personalities
	VoiceFemaleYoung   = "21m00Tcm4TlvDq8ikWAM" // Rachel - young female (default)
	VoiceFemaleMature  = "EXAVITQu4vr4xnSDxMaL" // Bella - mature female
	VoiceFemaleCool    = "MF3mGyEYCl7XYWbV9V6O" // Elli - cool, calm
	VoiceFemaleWarm    = "XrExE9yKIg1WjnnlVkGX" // Matilda - warm
	VoiceFemaleSoft    = "oWAxZDx7w5VEj9dCyTzz" // Grace - soft-spoken
	VoiceFemaleStrong  = "AZnzlk1XvdvUeBnXmlld" // Domi - strong, assertive
)

// VoiceGender represents the gender category for voice selection
type VoiceGender string

const (
	VoiceGenderMale   VoiceGender = "male"
	VoiceGenderFemale VoiceGender = "female"
)

// SelectVoiceForCharacter intelligently selects a voice based on character name and role
// Uses a hash of the name to ensure consistency (same name = same voice)
func SelectVoiceForCharacter(name string, role string) string {
	// Determine gender from name patterns and use hash for voice selection
	gender := determineGenderFromName(name)

	// Hash the name to get a consistent voice selection for this character
	hash := hashString(name)

	if gender == VoiceGenderMale {
		maleVoices := []string{
			VoiceMaleRough,
			VoiceMaleYoung,
			VoiceMaleCalm,
			VoiceMaleOld,
			VoiceMaleGruff,
			VoiceMaleDeep,
		}
		return maleVoices[hash%len(maleVoices)]
	} else {
		femaleVoices := []string{
			VoiceFemaleYoung,
			VoiceFemaleMature,
			VoiceFemaleCool,
			VoiceFemaleWarm,
			VoiceFemaleSoft,
			VoiceFemaleStrong,
		}
		return femaleVoices[hash%len(femaleVoices)]
	}
}

// determineGenderFromName makes an educated guess about gender based on name
// This is a simple heuristic - in reality, names are diverse and gender-neutral names exist
func determineGenderFromName(name string) VoiceGender {
	// Common name endings that suggest gender
	// This is a simplified approach - real implementation could use a name database

	// Split name into parts
	if len(name) == 0 {
		return VoiceGenderMale // default
	}

	// Simple heuristic: hash-based with slight bias
	// We want variety, so we'll use a simple rule:
	// - Names ending in 'a', 'e', 'i' are more likely female
	// - Otherwise, determine by hash for variety

	lastChar := name[len(name)-1]

	// Strong female indicators
	if lastChar == 'a' || lastChar == 'e' {
		return VoiceGenderFemale
	}

	// Use hash for everything else to ensure good distribution
	hash := hashString(name)
	if hash%2 == 0 {
		return VoiceGenderMale
	}
	return VoiceGenderFemale
}

// hashString creates a consistent hash for a string
func hashString(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}
