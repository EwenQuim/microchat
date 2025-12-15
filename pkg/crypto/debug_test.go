package crypto

import (
	"testing"
)

// TestActualPayload tests with the actual payload from the user
func TestActualPayload(t *testing.T) {
	// Actual data from the frontend
	content := "hi"
	room := "general" // Assuming this is the room name
	timestamp := int64(1765795113)
	pubkey := "02ed34ac037042436491f89080d96e1b93ca6a742d942488aec3081c5c5b80d669"
	signature := "c7d58aabf55f75be2de65c8cbb949188289246cf20565d1a8bdcc3cc1288411758b7346de623c65db44b99971c0bd2a762e409e4134e98a762c40daa0b10b987"

	// Try to verify
	err := VerifyMessageSignature(pubkey, signature, content, room, timestamp)
	if err != nil {
		t.Logf("Verification failed: %v", err)
		t.Logf("This will help us understand what's wrong")

		// Let's also create the event hash to see what we get
		eventHash := createEventHash(pubkey, timestamp, content, room)
		t.Logf("Backend event hash: %s", eventHash)
	} else {
		t.Logf("Verification succeeded!")
	}
}
