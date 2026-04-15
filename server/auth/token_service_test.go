package auth

import (
	"testing"
	"time"
)

func TestTokenService(t *testing.T) {
	expiry := 10 * time.Millisecond
	service := NewTokenServiceWithDuration(expiry)

	token := service.IssueToken()
	if err := service.AcceptToken(token); err != nil {
		t.Fatalf("accept valid token: %v", err)
	}

	if err := service.AcceptToken(token); err == nil {
		t.Fatalf("accept revoked token: expected error, got nil")
	}

	token = service.IssueToken()
	time.Sleep(expiry * 2)
	if err := service.AcceptToken(token); err == nil {
		t.Fatalf("accept expired token: expected error, got nil")
	}
}