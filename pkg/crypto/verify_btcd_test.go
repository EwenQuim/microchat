package crypto

import (
	"testing"
)

func TestVerifyBTCD(t *testing.T) {
	// Actual data from the frontend
	content := "hello"
	room := "general"
	timestamp := int64(1765796171)
	pubkey := "036903c174e82ef03e7fd5d721f233fa7b86eea298fda2e27372015b32d2bc7a29"
	signature := "18b5d24af7cf955e68cbbdfa111cd75ff7f3290eee1e6e73370a60d2591976464312bd757fd8a9fa6b915361bf6727acc62de7fc2f920ebab00a3465d9fe2ce7"

	err := VerifyMessageSignatureBTCD(pubkey, signature, content, room, timestamp)
	if err != nil {
		t.Errorf("BTCD verification failed: %v", err)
	} else {
		t.Log("BTCD verification SUCCESS!")
	}
}
