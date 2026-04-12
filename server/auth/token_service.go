package auth

type ITokenService interface {
	// IssueToken generates a new token for authentication
	IssueToken() (string, error)
	// RevokeToken invalidates a token
	RevokeToken(token string) error
	// AcceptToken validates a token and deletes it if valid (one-time use)
	AcceptToken(token string) error
}

func NewTokenService() (*ITokenService, error) {
	// TODO: Implement token service initialization logic
	return nil, nil
}