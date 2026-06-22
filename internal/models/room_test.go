package models

import (
	"context"
	"testing"
)

func strPtr(s string) *string { return &s }

func TestOutTransform_PasswordProtectedHidesLastMessage(t *testing.T) {
	r := Room{
		Name:                 "secret",
		HasPassword:          true,
		LastMessageContent:   strPtr("hello"),
		LastMessageUser:      strPtr("alice"),
		LastMessageTimestamp: strPtr("2026-06-22T00:00:00Z"),
	}

	if err := r.OutTransform(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.LastMessageContent != nil || r.LastMessageUser != nil || r.LastMessageTimestamp != nil {
		t.Fatal("password-protected room must not expose last message fields")
	}
}

func TestOutTransform_PublicKeepsLastMessage(t *testing.T) {
	r := Room{
		Name:                 "lobby",
		HasPassword:          false,
		LastMessageContent:   strPtr("hello"),
		LastMessageUser:      strPtr("alice"),
		LastMessageTimestamp: strPtr("2026-06-22T00:00:00Z"),
	}

	if err := r.OutTransform(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.LastMessageContent == nil || r.LastMessageUser == nil || r.LastMessageTimestamp == nil {
		t.Fatal("public room must keep its last message fields")
	}
}
