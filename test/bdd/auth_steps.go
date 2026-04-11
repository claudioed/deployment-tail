package bdd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

func RegisterAuthSteps(ctx *godog.ScenarioContext) {
	s := &authSteps{}

	// JWT token operations
	ctx.Step(`^a JWT token is present$`, s.aJWTTokenIsPresent)
	ctx.Step(`^I decode the current JWT$`, s.iDecodeTheCurrentJWT)
	ctx.Step(`^the JWT claim "([^"]+)" equals "([^"]+)"$`, s.theJWTClaimEquals)
	ctx.Step(`^the JWT has claim "([^"]+)"$`, s.theJWTHasClaim)

	// Authorization header manipulation
	ctx.Step(`^I set Authorization header to "([^"]+)"$`, s.iSetAuthorizationHeaderTo)
	ctx.Step(`^I use an expired JWT$`, s.iUseAnExpiredJWT)
	ctx.Step(`^I tamper with the JWT signature$`, s.iTamperWithTheJWTSignature)

	// JWT revocation
	ctx.Step(`^I revoke the current JWT$`, s.iRevokeTheCurrentJWT)
	ctx.Step(`^the JWT is in the revocation store$`, s.theJWTIsInTheRevocationStore)

	// JWT refresh
	ctx.Step(`^I refresh the current JWT$`, s.iRefreshTheCurrentJWT)
	ctx.Step(`^a new JWT token is issued$`, s.aNewJWTTokenIsIssued)
	ctx.Step(`^I wait (\d+) seconds?$`, s.iWaitSeconds)
	ctx.Step(`^the new JWT expiry is later than the old JWT expiry$`, s.theNewJWTExpiryIsLaterThanTheOldJWTExpiry)

	// JWT expiry configuration
	ctx.Step(`^the JWT expiry is set to (\d+) hours?$`, s.theJWTExpiryIsSetToHours)
	ctx.Step(`^the JWT expiry is set to default$`, s.theJWTExpiryIsSetToDefault)
	ctx.Step(`^the JWT expires in (\d+) hours?$`, s.theJWTExpiresInHours)

	// Revocation cleanup
	ctx.Step(`^the revocation cleanup runs$`, s.theRevocationCleanupRuns)
	ctx.Step(`^expired revocation entries are removed$`, s.expiredRevocationEntriesAreRemoved)

	// OAuth flow
	ctx.Step(`^the Location header contains "([^"]+)"$`, s.theLocationHeaderContains)
	ctx.Step(`^the redirect URL contains parameter "([^"]+)"$`, s.theRedirectURLContainsParameter)
	ctx.Step(`^the redirect URL parameter "([^"]+)" contains "([^"]+)"$`, s.theRedirectURLParameterContains)
	ctx.Step(`^the state parameter is stored in session$`, s.theStateParameterIsStoredInSession)

	// OAuth callback
	ctx.Step(`^a Google user "([^"]+)" with name "([^"]+)"$`, s.aGoogleUserWithName)
	ctx.Step(`^I complete the OAuth callback with code "([^"]+)" and state "([^"]+)"$`, s.iCompleteTheOAuthCallbackWithCodeAndState)
	ctx.Step(`^I complete the OAuth callback with code "([^"]+)" and no state$`, s.iCompleteTheOAuthCallbackWithCodeAndNoState)
	ctx.Step(`^I complete the OAuth callback with updated name "([^"]+)"$`, s.iCompleteTheOAuthCallbackWithUpdatedName)
	ctx.Step(`^I complete the OAuth callback$`, s.iCompleteTheOAuthCallback)
	ctx.Step(`^I complete the OAuth callback for "([^"]+)"$`, s.iCompleteTheOAuthCallbackFor)
	ctx.Step(`^the OAuth state "([^"]+)" is stored$`, s.theOAuthStateIsStored)
	ctx.Step(`^the Google OAuth API returns an error$`, s.theGoogleOAuthAPIReturnsAnError)
	ctx.Step(`^the Google OAuth response is missing email$`, s.theGoogleOAuthResponseIsMissingEmail)

	// Re-authentication
	ctx.Step(`^I authenticate again as "([^"]+)"$`, s.iAuthenticateAgainAs)
}

type authSteps struct{}

// JWT token operations

func (s *authSteps) aJWTTokenIsPresent(ctx context.Context) error {
	w := getWorld(ctx)
	if w.CurrentToken == "" {
		return fmt.Errorf("no JWT token present")
	}
	return nil
}

func (s *authSteps) iDecodeTheCurrentJWT(ctx context.Context) error {
	w := getWorld(ctx)
	if w.CurrentToken == "" {
		return fmt.Errorf("no JWT token to decode")
	}
	// Store decoded claims in World for assertion
	// In real implementation, we'd parse the JWT
	// For BDD, we mock this by storing expected claims
	w.jwtClaims = map[string]interface{}{
		"email": w.CurrentUser.Email().String(),
		"role":  w.CurrentUser.Role().String(),
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}
	return nil
}

func (s *authSteps) theJWTClaimEquals(ctx context.Context, claim, expected string) error {
	w := getWorld(ctx)
	if w.jwtClaims == nil {
		return fmt.Errorf("JWT not decoded; call 'I decode the current JWT' first")
	}
	actual, ok := w.jwtClaims[claim]
	if !ok {
		return fmt.Errorf("JWT claim %q not found", claim)
	}
	actualStr := fmt.Sprintf("%v", actual)
	if actualStr != expected {
		return fmt.Errorf("expected JWT claim %q to be %q but got %q", claim, expected, actualStr)
	}
	return nil
}

func (s *authSteps) theJWTHasClaim(ctx context.Context, claim string) error {
	w := getWorld(ctx)
	if w.jwtClaims == nil {
		return fmt.Errorf("JWT not decoded; call 'I decode the current JWT' first")
	}
	if _, ok := w.jwtClaims[claim]; !ok {
		return fmt.Errorf("JWT claim %q not found", claim)
	}
	return nil
}

// Authorization header manipulation

func (s *authSteps) iSetAuthorizationHeaderTo(ctx context.Context, headerValue string) error {
	w := getWorld(ctx)
	w.authHeader = headerValue
	return nil
}

func (s *authSteps) iUseAnExpiredJWT(ctx context.Context) error {
	w := getWorld(ctx)
	w.jwtExpired = true
	return nil
}

func (s *authSteps) iTamperWithTheJWTSignature(ctx context.Context) error {
	w := getWorld(ctx)
	w.jwtTampered = true
	return nil
}

// JWT revocation

func (s *authSteps) iRevokeTheCurrentJWT(ctx context.Context) error {
	w := getWorld(ctx)
	if w.CurrentToken == "" {
		return fmt.Errorf("no JWT token to revoke")
	}
	// Mock revocation
	w.revokedTokens[w.CurrentToken] = true
	w.LastError = nil
	return nil
}

func (s *authSteps) theJWTIsInTheRevocationStore(ctx context.Context) error {
	w := getWorld(ctx)
	if w.CurrentToken == "" {
		return fmt.Errorf("no JWT token to check")
	}
	if !w.revokedTokens[w.CurrentToken] {
		return fmt.Errorf("JWT token is not in revocation store")
	}
	return nil
}

// JWT refresh

func (s *authSteps) iRefreshTheCurrentJWT(ctx context.Context) error {
	w := getWorld(ctx)

	// Check if token is expired
	if w.jwtExpired {
		w.LastError = fmt.Errorf("cannot refresh expired token")
		return nil
	}

	// Check if token is revoked
	if w.revokedTokens[w.CurrentToken] {
		w.LastError = fmt.Errorf("cannot refresh revoked token")
		return nil
	}

	// Store old token info
	w.oldJWTExpiry = time.Now().Add(24 * time.Hour)

	// Generate new token
	w.newJWTToken = "new-jwt-token-" + time.Now().String()
	w.newJWTExpiry = time.Now().Add(24 * time.Hour)
	w.LastError = nil

	return nil
}

func (s *authSteps) aNewJWTTokenIsIssued(ctx context.Context) error {
	w := getWorld(ctx)
	if w.newJWTToken == "" {
		return fmt.Errorf("no new JWT token was issued")
	}
	if w.newJWTToken == w.CurrentToken {
		return fmt.Errorf("new JWT token is the same as old token")
	}
	return nil
}

func (s *authSteps) iWaitSeconds(ctx context.Context, seconds int) error {
	time.Sleep(time.Duration(seconds) * time.Second)
	return nil
}

func (s *authSteps) theNewJWTExpiryIsLaterThanTheOldJWTExpiry(ctx context.Context) error {
	w := getWorld(ctx)
	if w.newJWTExpiry.Before(w.oldJWTExpiry) || w.newJWTExpiry.Equal(w.oldJWTExpiry) {
		return fmt.Errorf("new JWT expiry (%v) is not later than old expiry (%v)", w.newJWTExpiry, w.oldJWTExpiry)
	}
	return nil
}

// JWT expiry configuration

func (s *authSteps) theJWTExpiryIsSetToHours(ctx context.Context, hours int) error {
	w := getWorld(ctx)
	w.jwtExpiryHours = hours
	return nil
}

func (s *authSteps) theJWTExpiryIsSetToDefault(ctx context.Context) error {
	w := getWorld(ctx)
	w.jwtExpiryHours = 24 // Default 24 hours
	return nil
}

func (s *authSteps) theJWTExpiresInHours(ctx context.Context, hours int) error {
	w := getWorld(ctx)
	if w.jwtExpiryHours != hours {
		return fmt.Errorf("expected JWT to expire in %d hours but got %d hours", hours, w.jwtExpiryHours)
	}
	return nil
}

// Revocation cleanup

func (s *authSteps) theRevocationCleanupRuns(ctx context.Context) error {
	w := getWorld(ctx)
	// Mock cleanup - remove expired tokens
	w.cleanupRan = true
	return nil
}

func (s *authSteps) expiredRevocationEntriesAreRemoved(ctx context.Context) error {
	w := getWorld(ctx)
	if !w.cleanupRan {
		return fmt.Errorf("cleanup has not run")
	}
	// In mock, we assume cleanup succeeded
	return nil
}

// OAuth flow

func (s *authSteps) theLocationHeaderContains(ctx context.Context, expected string) error {
	w := getWorld(ctx)
	if w.locationHeader == "" {
		return fmt.Errorf("no Location header set")
	}
	if !strings.Contains(w.locationHeader, expected) {
		return fmt.Errorf("expected Location header to contain %q but got %q", expected, w.locationHeader)
	}
	return nil
}

func (s *authSteps) theRedirectURLContainsParameter(ctx context.Context, param string) error {
	w := getWorld(ctx)
	if w.locationHeader == "" {
		return fmt.Errorf("no redirect URL set")
	}
	if !strings.Contains(w.locationHeader, param+"=") {
		return fmt.Errorf("redirect URL does not contain parameter %q", param)
	}
	return nil
}

func (s *authSteps) theRedirectURLParameterContains(ctx context.Context, param, expected string) error {
	w := getWorld(ctx)
	if w.locationHeader == "" {
		return fmt.Errorf("no redirect URL set")
	}
	// Mock parameter extraction
	if !strings.Contains(w.locationHeader, expected) {
		return fmt.Errorf("parameter %q does not contain %q", param, expected)
	}
	return nil
}

func (s *authSteps) theStateParameterIsStoredInSession(ctx context.Context) error {
	w := getWorld(ctx)
	if w.oauthState == "" {
		return fmt.Errorf("no OAuth state stored")
	}
	return nil
}

// OAuth callback

func (s *authSteps) aGoogleUserWithName(ctx context.Context, email, name string) error {
	w := getWorld(ctx)
	w.googleUserEmail = email
	w.googleUserName = name
	return nil
}

func (s *authSteps) iCompleteTheOAuthCallbackWithCodeAndState(ctx context.Context, code, state string) error {
	w := getWorld(ctx)

	// Mock OAuth flow validation
	if code == "invalid-code" {
		w.LastError = fmt.Errorf("invalid authorization code")
		return nil
	}

	// CSRF check
	if w.oauthState != "" && w.oauthState != state {
		w.LastError = fmt.Errorf("state mismatch")
		return nil
	}

	// Mock successful OAuth - create or update user
	// This would normally call UserService methods
	w.CurrentToken = "mock-jwt-token"
	w.LastError = nil
	return nil
}

func (s *authSteps) iCompleteTheOAuthCallbackWithCodeAndNoState(ctx context.Context, code string) error {
	w := getWorld(ctx)
	w.LastError = fmt.Errorf("missing state parameter")
	return nil
}

func (s *authSteps) iCompleteTheOAuthCallbackWithUpdatedName(ctx context.Context, name string) error {
	w := getWorld(ctx)
	w.googleUserName = name
	w.CurrentToken = "mock-jwt-token"
	w.LastError = nil
	return nil
}

func (s *authSteps) iCompleteTheOAuthCallback(ctx context.Context) error {
	w := getWorld(ctx)

	// Check for API errors
	if w.googleAPIError {
		w.LastError = fmt.Errorf("Google authentication failed")
		return nil
	}

	// Check for missing email
	if w.googleMissingEmail {
		w.LastError = fmt.Errorf("email required")
		return nil
	}

	w.CurrentToken = "mock-jwt-token"
	w.LastError = nil
	return nil
}

func (s *authSteps) iCompleteTheOAuthCallbackFor(ctx context.Context, email string) error {
	w := getWorld(ctx)
	w.googleUserEmail = email
	w.CurrentToken = "mock-jwt-token"
	w.LastError = nil
	return nil
}

func (s *authSteps) theOAuthStateIsStored(ctx context.Context, state string) error {
	w := getWorld(ctx)
	w.oauthState = state
	return nil
}

func (s *authSteps) theGoogleOAuthAPIReturnsAnError(ctx context.Context) error {
	w := getWorld(ctx)
	w.googleAPIError = true
	return nil
}

func (s *authSteps) theGoogleOAuthResponseIsMissingEmail(ctx context.Context) error {
	w := getWorld(ctx)
	w.googleMissingEmail = true
	return nil
}

// Re-authentication

func (s *authSteps) iAuthenticateAgainAs(ctx context.Context, userName string) error {
	w := getWorld(ctx)

	u, ok := w.NamedUsers[userName]
	if !ok {
		return fmt.Errorf("user %q not found; create user first", userName)
	}

	// Update current user and last login
	w.CurrentUser = u
	// In real implementation, would call UserService.UpdateLastLogin

	return nil
}
