package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/yi-tech/go-user-service/internal/domain/auth"
	"go.uber.org/zap"
)

const (
	sessionKeyPrefix     = "session:"
	refreshTokenKeyPrefix = "refresh:"
	sessionExpiry        = 24 * 7 * time.Hour // 1 week
)

// authRepository implements the domain auth.Repository interface
type authRepository struct {
	redis  *redis.Client
	logger *zap.Logger
}

// NewAuthRepository creates a new authentication repository
func NewAuthRepository(redis *redis.Client, logger *zap.Logger) auth.Repository {
	return &authRepository{
		redis:  redis,
		logger: logger,
	}
}

// CreateSession stores a new session
func (r *authRepository) CreateSession(ctx context.Context, session *auth.Session) error {
	// Serialize session to JSON
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		r.logger.Error("Failed to marshal session", zap.Error(err))
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Store session by ID
	sessionKey := fmt.Sprintf("%s%s", sessionKeyPrefix, session.ID)
	if err := r.redis.Set(ctx, sessionKey, sessionJSON, sessionExpiry).Err(); err != nil {
		r.logger.Error("Failed to store session", zap.Error(err), zap.String("session_id", session.ID))
		return fmt.Errorf("failed to store session: %w", err)
	}

	// Store mapping from refresh token to session ID
	refreshKey := fmt.Sprintf("%s%s", refreshTokenKeyPrefix, session.RefreshToken)
	if err := r.redis.Set(ctx, refreshKey, session.ID, sessionExpiry).Err(); err != nil {
		r.logger.Error("Failed to store refresh token mapping", zap.Error(err))
		// Try to clean up the session if refresh token mapping fails
		_ = r.redis.Del(ctx, sessionKey)
		return fmt.Errorf("failed to store refresh token mapping: %w", err)
	}

	return nil
}

// GetSessionByID retrieves a session by ID
func (r *authRepository) GetSessionByID(ctx context.Context, id string) (*auth.Session, error) {
	sessionKey := fmt.Sprintf("%s%s", sessionKeyPrefix, id)
	
	// Get session from Redis
	sessionJSON, err := r.redis.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		r.logger.Error("Failed to get session", zap.Error(err), zap.String("session_id", id))
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Deserialize session
	var session auth.Session
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		r.logger.Error("Failed to unmarshal session", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// GetSessionByRefreshToken retrieves a session by refresh token
func (r *authRepository) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*auth.Session, error) {
	refreshKey := fmt.Sprintf("%s%s", refreshTokenKeyPrefix, refreshToken)
	
	// Get session ID from refresh token
	sessionID, err := r.redis.Get(ctx, refreshKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		r.logger.Error("Failed to get session ID from refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to get session ID: %w", err)
	}

	// Get session by ID
	return r.GetSessionByID(ctx, sessionID)
}

// DeleteSession removes a session
func (r *authRepository) DeleteSession(ctx context.Context, id string) error {
	// Get session first to get the refresh token
	session, err := r.GetSessionByID(ctx, id)
	if err != nil {
		return err
	}

	if session == nil {
		return nil
	}

	// Delete session
	sessionKey := fmt.Sprintf("%s%s", sessionKeyPrefix, id)
	if err := r.redis.Del(ctx, sessionKey).Err(); err != nil {
		r.logger.Error("Failed to delete session", zap.Error(err), zap.String("session_id", id))
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Delete refresh token mapping
	refreshKey := fmt.Sprintf("%s%s", refreshTokenKeyPrefix, session.RefreshToken)
	if err := r.redis.Del(ctx, refreshKey).Err(); err != nil {
		r.logger.Error("Failed to delete refresh token mapping", zap.Error(err))
		// Non-critical error, we've already deleted the session
	}

	return nil
}

// DeleteUserSessions removes all sessions for a user
func (r *authRepository) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	// This is a simplified implementation that would need to be enhanced in a production system
	// In a real system, we would need to maintain a secondary index of sessions by user ID
	// For now, we'll just log that this operation is not fully implemented
	r.logger.Warn("DeleteUserSessions is not fully implemented", zap.String("user_id", userID.String()))
	
	// In a real implementation, we would:
	// 1. Find all sessions for the user
	// 2. Delete each session and its refresh token mapping
	
	return nil
}
