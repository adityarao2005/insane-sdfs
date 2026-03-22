package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

var (
	ErrInviteNotFound = errors.New("invite token not found")
	ErrInviteExpired  = errors.New("invite token expired")
	ErrInviteUsed     = errors.New("invite token already used")
)

type inviteRecord struct {
	ExpiresAt time.Time
	Used      bool
}

type InviteTokenStore struct {
	mu     sync.Mutex
	tokens map[[32]byte]inviteRecord
}

func NewInviteTokenStore() *InviteTokenStore {
	return &InviteTokenStore{tokens: make(map[[32]byte]inviteRecord)}
}

// Issue creates a single-use invite token and stores only its hash.
func (s *InviteTokenStore) Issue(ttl time.Duration, now time.Time) (token string, expiresAt time.Time, err error) {
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", time.Time{}, err
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	expiresAt = now.Add(ttl)
	h := sha256.Sum256([]byte(token))

	s.mu.Lock()
	s.tokens[h] = inviteRecord{ExpiresAt: expiresAt, Used: false}
	s.mu.Unlock()

	return token, expiresAt, nil
}

// Consume validates and consumes a token exactly once.
func (s *InviteTokenStore) Consume(token string, now time.Time) error {
	h := sha256.Sum256([]byte(token))

	s.mu.Lock()
	defer s.mu.Unlock()

	rec, ok := s.tokens[h]
	if !ok {
		return ErrInviteNotFound
	}
	if now.After(rec.ExpiresAt) {
		delete(s.tokens, h)
		return ErrInviteExpired
	}
	if rec.Used {
		return ErrInviteUsed
	}
	rec.Used = true
	s.tokens[h] = rec
	return nil
}
