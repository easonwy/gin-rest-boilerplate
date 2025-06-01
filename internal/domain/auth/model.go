package auth

import (
	"time"

	"github.com/google/uuid"
)

// TokenPair represents an access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Session represents a user authentication session
type Session struct {
	ID           string    `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	RefreshToken string    `json:"-"` // Never expose in JSON
	UserAgent    string    `json:"user_agent"`
	ClientIP     string    `json:"client_ip"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// NewSession creates a new user session
func NewSession(userID uuid.UUID, refreshToken, userAgent, clientIP string, expiry time.Duration) *Session {
	return &Session{
		ID:           uuid.New().String(),
		UserID:       userID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		ClientIP:     clientIP,
		ExpiresAt:    time.Now().Add(expiry),
		CreatedAt:    time.Now(),
	}
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
