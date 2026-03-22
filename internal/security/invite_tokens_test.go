package security

import (
	"testing"
	"time"
)

func TestInviteSingleUse(t *testing.T) {
	store := NewInviteTokenStore()
	now := time.Unix(1000, 0).UTC()
	tok, _, err := store.Issue(10*time.Minute, now)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	if err := store.Consume(tok, now.Add(time.Minute)); err != nil {
		t.Fatalf("first consume: %v", err)
	}
	if err := store.Consume(tok, now.Add(2*time.Minute)); err != ErrInviteUsed {
		t.Fatalf("expected ErrInviteUsed, got: %v", err)
	}
}

func TestInviteExpiry(t *testing.T) {
	store := NewInviteTokenStore()
	now := time.Unix(1000, 0).UTC()
	tok, _, err := store.Issue(1*time.Minute, now)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	if err := store.Consume(tok, now.Add(2*time.Minute)); err != ErrInviteExpired {
		t.Fatalf("expected ErrInviteExpired, got: %v", err)
	}
}
