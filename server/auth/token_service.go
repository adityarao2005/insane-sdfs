package auth

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

type ITokenService interface {
	// IssueToken generates a new token for authentication
	// revoke it after 10 mins
	IssueToken() string
	// AcceptToken validates a token and deletes it if valid (one-time use)
	AcceptToken(token string) error
}

type TokenService struct {
	mutex sync.Mutex
	tokens map[string]bool // Map of token to its status (e.g., "valid", "revoked")
	expiryDuration time.Duration
}

func (s *TokenService) IssueToken() string {

	// generate a random token
	token := rand.Text()

	// Start a goroutine to revoke the token after 10 minutes
	go func() {
		// Revoke the token after 10 minutes
		time.Sleep(s.expiryDuration)
		
		s.mutex.Lock()
		defer s.mutex.Unlock()
		delete(s.tokens, token)
	}()

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.tokens[token] = true // Mark token as valid

	return token
}

func (s *TokenService) AcceptToken(token string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, exists := s.tokens[token]

	if !exists {
		return fmt.Errorf("invalid or expired token")
	}

	delete(s.tokens, token)
	return nil
}

func NewTokenService() ITokenService {
	return &TokenService{
		mutex: sync.Mutex{},
		tokens: make(map[string]bool),
		expiryDuration: 10 * time.Minute,
	}
}

func NewTokenServiceWithDuration(expiry time.Duration) ITokenService {
	return &TokenService{
		mutex: sync.Mutex{},
		tokens: make(map[string]bool),
		expiryDuration: expiry,
	}
}