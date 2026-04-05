package jwt

import (
	"context"
	"testing"
	"time"
)

func TestRevocationStore_IsRevoked(t *testing.T) {
	store := NewRevocationStore(nil) // No DB for unit test

	tokenHash := "test-hash-123"
	futureExpiry := time.Now().Add(1 * time.Hour)

	// Initially not revoked
	if store.IsRevoked(tokenHash) {
		t.Error("Token should not be revoked initially")
	}

	// Add to blacklist
	store.blacklist.Store(tokenHash, futureExpiry)

	// Now should be revoked
	if !store.IsRevoked(tokenHash) {
		t.Error("Token should be revoked after adding to blacklist")
	}
}

func TestRevocationStore_IsRevoked_Expired(t *testing.T) {
	store := NewRevocationStore(nil)

	tokenHash := "expired-token"
	pastExpiry := time.Now().Add(-1 * time.Hour)

	// Add expired token
	store.blacklist.Store(tokenHash, pastExpiry)

	// Should not be considered revoked (expired tokens are auto-removed)
	if store.IsRevoked(tokenHash) {
		t.Error("Expired token should not be considered revoked")
	}

	// Should be removed from blacklist
	if _, exists := store.blacklist.Load(tokenHash); exists {
		t.Error("Expired token should be removed from blacklist")
	}
}

func TestRevocationStore_GetBlacklistSize(t *testing.T) {
	store := NewRevocationStore(nil)

	if store.GetBlacklistSize() != 0 {
		t.Error("Initial blacklist size should be 0")
	}

	futureExpiry := time.Now().Add(1 * time.Hour)
	store.blacklist.Store("token1", futureExpiry)
	store.blacklist.Store("token2", futureExpiry)
	store.blacklist.Store("token3", futureExpiry)

	if store.GetBlacklistSize() != 3 {
		t.Errorf("Expected blacklist size 3, got %d", store.GetBlacklistSize())
	}
}

func TestRevocationStore_Cleanup_Memory(t *testing.T) {
	store := NewRevocationStore(nil)

	now := time.Now()
	pastExpiry := now.Add(-1 * time.Hour)
	futureExpiry := now.Add(1 * time.Hour)

	// Add both expired and valid tokens
	store.blacklist.Store("expired1", pastExpiry)
	store.blacklist.Store("expired2", pastExpiry)
	store.blacklist.Store("valid1", futureExpiry)
	store.blacklist.Store("valid2", futureExpiry)

	if store.GetBlacklistSize() != 4 {
		t.Errorf("Expected 4 tokens before cleanup, got %d", store.GetBlacklistSize())
	}

	// Run cleanup
	ctx := context.Background()
	store.cleanup(ctx)

	// Only valid tokens should remain
	if store.GetBlacklistSize() != 2 {
		t.Errorf("Expected 2 tokens after cleanup, got %d", store.GetBlacklistSize())
	}

	// Verify expired tokens are gone
	if _, exists := store.blacklist.Load("expired1"); exists {
		t.Error("Expired token should be removed")
	}

	// Verify valid tokens remain
	if _, exists := store.blacklist.Load("valid1"); !exists {
		t.Error("Valid token should remain")
	}
}

func TestRevocationStore_Stop(t *testing.T) {
	store := NewRevocationStore(nil)

	// Start and immediately stop
	store.Stop()

	// Check that stopCh is closed
	select {
	case <-store.stopCh:
		// Channel is closed, as expected
	default:
		t.Error("Stop channel should be closed")
	}
}

func TestRevocationStore_ConcurrentAccess(t *testing.T) {
	store := NewRevocationStore(nil)

	futureExpiry := time.Now().Add(1 * time.Hour)

	// Simulate concurrent access
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			store.blacklist.Store("token", futureExpiry)
			time.Sleep(1 * time.Microsecond)
		}
		done <- true
	}()

	// Reader goroutines
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				store.IsRevoked("token")
				time.Sleep(1 * time.Microsecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 6; i++ {
		<-done
	}

	// If we get here without panic, concurrent access works
	t.Log("Concurrent access test passed")
}

func TestRevocationStore_MultipleTokens(t *testing.T) {
	store := NewRevocationStore(nil)

	futureExpiry := time.Now().Add(1 * time.Hour)

	// Add multiple tokens
	tokens := []string{"token1", "token2", "token3", "token4", "token5"}
	for _, token := range tokens {
		store.blacklist.Store(token, futureExpiry)
	}

	// Verify all are revoked
	for _, token := range tokens {
		if !store.IsRevoked(token) {
			t.Errorf("Token %s should be revoked", token)
		}
	}

	// Verify non-existent token is not revoked
	if store.IsRevoked("non-existent") {
		t.Error("Non-existent token should not be revoked")
	}
}
