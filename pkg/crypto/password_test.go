package crypto

import (
	"errors"
	"testing"
)

func TestHashPassword_Empty(t *testing.T) {
	if _, err := HashPassword(""); err == nil {
		t.Fatal("expected error for empty password")
	}
}

func TestHashPassword_RoundTrip(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if hash == "correct horse battery staple" {
		t.Fatal("hash must not equal the plaintext password")
	}
	if err := VerifyPassword("correct horse battery staple", hash); err != nil {
		t.Fatalf("verify of correct password should succeed: %v", err)
	}
}

func TestVerifyPassword_Wrong(t *testing.T) {
	hash, err := HashPassword("right")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = VerifyPassword("wrong", hash)
	if !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestVerifyPassword_EmptyInputs(t *testing.T) {
	if err := VerifyPassword("", "somehash"); err == nil {
		t.Fatal("expected error for empty password")
	}
	if err := VerifyPassword("pw", ""); err == nil {
		t.Fatal("expected error for empty hash")
	}
}
