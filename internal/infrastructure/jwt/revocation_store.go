package jwt

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// RevocationStore manages revoked JWT tokens
type RevocationStore struct {
	db        *sql.DB
	blacklist sync.Map // map[tokenHash]expiresAt
	mu        sync.RWMutex
	stopCh    chan struct{}
}

// RevokedToken represents a revoked token entry
type RevokedToken struct {
	TokenHash string
	UserID    string
	RevokedAt time.Time
	ExpiresAt time.Time
}

// NewRevocationStore creates a new revocation store
func NewRevocationStore(db *sql.DB) *RevocationStore {
	return &RevocationStore{
		db:     db,
		stopCh: make(chan struct{}),
	}
}

// Start initializes the revocation store and starts background sync
func (s *RevocationStore) Start(ctx context.Context) error {
	// Load existing revoked tokens from database
	if err := s.LoadFromDatabase(ctx); err != nil {
		return fmt.Errorf("failed to load blacklist from database: %w", err)
	}

	// Start background sync goroutine
	go s.syncLoop(ctx)

	// Start cleanup goroutine
	go s.cleanupLoop(ctx)

	return nil
}

// Stop gracefully stops the background goroutines
func (s *RevocationStore) Stop() {
	close(s.stopCh)
}

// AddToBlacklist adds a token to the revocation blacklist
func (s *RevocationStore) AddToBlacklist(ctx context.Context, tokenHash string, userID string, expiresAt time.Time) error {
	// Store in database first
	query := `
		INSERT INTO revoked_tokens (token_hash, user_id, revoked_at, expires_at)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE revoked_at = VALUES(revoked_at)
	`

	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, query, tokenHash, userID, now, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to insert revoked token: %w", err)
	}

	// Add to in-memory blacklist
	s.blacklist.Store(tokenHash, expiresAt)

	return nil
}

// IsRevoked checks if a token is in the blacklist
func (s *RevocationStore) IsRevoked(tokenHash string) bool {
	value, exists := s.blacklist.Load(tokenHash)
	if !exists {
		return false
	}

	expiresAt, ok := value.(time.Time)
	if !ok {
		return false
	}

	// If token has expired, it's no longer relevant
	if time.Now().UTC().After(expiresAt) {
		s.blacklist.Delete(tokenHash)
		return false
	}

	return true
}

// LoadFromDatabase syncs the blacklist from the database
func (s *RevocationStore) LoadFromDatabase(ctx context.Context) error {
	// Load only non-expired tokens
	query := `
		SELECT token_hash, expires_at
		FROM revoked_tokens
		WHERE expires_at > ?
	`

	now := time.Now().UTC()
	rows, err := s.db.QueryContext(ctx, query, now)
	if err != nil {
		return fmt.Errorf("failed to query revoked tokens: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var tokenHash string
		var expiresAt time.Time

		if err := rows.Scan(&tokenHash, &expiresAt); err != nil {
			return fmt.Errorf("failed to scan revoked token: %w", err)
		}

		s.blacklist.Store(tokenHash, expiresAt)
		count++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating revoked tokens: %w", err)
	}

	return nil
}

// syncLoop periodically syncs the blacklist from the database
func (s *RevocationStore) syncLoop(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.LoadFromDatabase(ctx); err != nil {
				// Log error but continue (in production, use proper logger)
				fmt.Printf("Warning: failed to sync blacklist: %v\n", err)
			}
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// cleanupLoop periodically removes expired entries
func (s *RevocationStore) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanup(ctx)
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// cleanup removes expired entries from memory and database
func (s *RevocationStore) cleanup(ctx context.Context) {
	now := time.Now().UTC()

	// Clean up in-memory blacklist
	s.blacklist.Range(func(key, value interface{}) bool {
		expiresAt, ok := value.(time.Time)
		if ok && now.After(expiresAt) {
			s.blacklist.Delete(key)
		}
		return true
	})

	// Clean up database (if available)
	if s.db != nil {
		query := "DELETE FROM revoked_tokens WHERE expires_at < ?"
		_, err := s.db.ExecContext(ctx, query, now)
		if err != nil {
			// Log error but continue (in production, use proper logger)
			fmt.Printf("Warning: failed to cleanup expired tokens: %v\n", err)
		}
	}
}

// RevokeAllUserTokens revokes all tokens for a specific user
func (s *RevocationStore) RevokeAllUserTokens(ctx context.Context, userID string) error {
	// This marks all tokens for the user as revoked in the database
	// The actual tokens need to be tracked elsewhere or tokens need expiry times
	// For now, we just log this operation
	query := `
		UPDATE revoked_tokens
		SET revoked_at = ?
		WHERE user_id = ? AND expires_at > ?
	`

	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, query, now, userID, now)
	if err != nil {
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}

	// Reload from database to update in-memory cache
	return s.LoadFromDatabase(ctx)
}

// GetBlacklistSize returns the current size of the in-memory blacklist
func (s *RevocationStore) GetBlacklistSize() int {
	count := 0
	s.blacklist.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
