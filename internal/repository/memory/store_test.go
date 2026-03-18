package memory

import (
	"context"
	"testing"
	"time"

	"github.com/EwenQuim/microchat/internal/services"
)

// saveAt saves a message then forcibly sets its timestamp to ts.
func saveAt(t *testing.T, s *Store, room string, ts time.Time) {
	t.Helper()
	ctx := context.Background()
	_, err := s.SaveMessage(ctx, room, "user", "content", "", "", 0)
	if err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}
	s.mu.Lock()
	msgs := s.messages[room]
	msgs[len(msgs)-1].Timestamp = ts
	s.mu.Unlock()
}

func TestGetMessages_DefaultLimit50(t *testing.T) {
	s := NewStore()
	ctx := context.Background()
	room := "testroom"
	base := time.Now().Add(-100 * time.Second)

	for i := range 80 {
		saveAt(t, s, room, base.Add(time.Duration(i)*time.Second))
	}

	msgs, err := s.GetMessages(ctx, room, services.MessageQueryParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 50 {
		t.Errorf("got %d messages, want 50 (default limit)", len(msgs))
	}
}

func TestGetMessages_LimitApplied(t *testing.T) {
	s := NewStore()
	ctx := context.Background()
	room := "testroom"
	base := time.Now().Add(-100 * time.Second)

	for i := range 80 {
		saveAt(t, s, room, base.Add(time.Duration(i)*time.Second))
	}

	msgs, err := s.GetMessages(ctx, room, services.MessageQueryParams{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 10 {
		t.Errorf("got %d messages, want 10", len(msgs))
	}
}

func TestGetMessages_BeforeCursor(t *testing.T) {
	s := NewStore()
	ctx := context.Background()
	room := "testroom"
	base := time.Now().Add(-200 * time.Second)

	for i := range 20 {
		saveAt(t, s, room, base.Add(time.Duration(i)*time.Second))
	}

	// before = base + 10s → messages 0..9 qualify (10 messages, timestamps 0s–9s)
	before := base.Add(10 * time.Second)
	msgs, err := s.GetMessages(ctx, room, services.MessageQueryParams{Limit: 50, Before: &before})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 10 {
		t.Errorf("got %d messages before cursor, want 10", len(msgs))
	}
}

func TestGetMessages_AscendingOrder(t *testing.T) {
	s := NewStore()
	ctx := context.Background()
	room := "testroom"
	base := time.Now().Add(-100 * time.Second)

	for i := range 5 {
		saveAt(t, s, room, base.Add(time.Duration(i)*time.Second))
	}

	msgs, err := s.GetMessages(ctx, room, services.MessageQueryParams{Limit: 50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 1; i < len(msgs); i++ {
		if msgs[i].Timestamp.Before(msgs[i-1].Timestamp) {
			t.Errorf("messages not in ascending order at index %d", i)
		}
	}
}

func TestGetMessages_EmptyRoom(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	msgs, err := s.GetMessages(ctx, "noroom", services.MessageQueryParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("got %d messages for empty room, want 0", len(msgs))
	}
}
