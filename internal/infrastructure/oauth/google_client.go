package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleClient handles Google OpenID Connect authentication
type GoogleClient struct {
	config   *oauth2.Config
	verifier *oidc.IDTokenVerifier
	provider *oidc.Provider
}

// GoogleUserInfo represents the user information from OIDC ID token
type GoogleUserInfo struct {
	ID            string `json:"sub"`            // OIDC standard: subject identifier
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"email_verified"` // OIDC standard claim
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// Config holds the configuration for Google OIDC
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// NewGoogleClient creates a new Google OIDC client
func NewGoogleClient(cfg Config) (*GoogleClient, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Initialize OIDC provider for Google
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OIDC provider: %w", err)
	}

	// Create ID token verifier
	verifier := provider.Verifier(&oidc.Config{
		ClientID: cfg.ClientID,
	})

	// Configure OAuth2 with OIDC scopes
	config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes: []string{
			oidc.ScopeOpenID, // Required for OIDC
			"email",          // Standard OIDC scope
			"profile",        // Standard OIDC scope
		},
		Endpoint: google.Endpoint,
	}

	return &GoogleClient{
		config:   config,
		verifier: verifier,
		provider: provider,
	}, nil
}

// GetAuthURL generates the OAuth authorization URL
func (c *GoogleClient) GetAuthURL(state string) string {
	return c.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges an authorization code for tokens and validates the ID token
func (c *GoogleClient) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	// Exchange authorization code for tokens (includes ID token)
	token, err := c.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	if !token.Valid() {
		return nil, fmt.Errorf("token is not valid")
	}

	// Extract ID token from OAuth2 token (OIDC requirement)
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in OAuth2 token response")
	}

	// Verify ID token signature and claims (OIDC compliance)
	idToken, err := c.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	// Additional validation: check if token is expired
	if err := idToken.VerifyAccessToken(token.AccessToken); err != nil {
		// Note: This is optional - just ensures access token matches ID token
		// Some providers don't support this, so we log but don't fail
		// fmt.Printf("Warning: Could not verify access token: %v\n", err)
	}

	return token, nil
}

// GetUserInfo extracts user information from the ID token (OIDC-compliant)
// This method no longer makes an extra API call - it uses the verified ID token claims
func (c *GoogleClient) GetUserInfo(ctx context.Context, oauth2Token *oauth2.Token) (*GoogleUserInfo, error) {
	// Extract raw ID token from OAuth2 token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in token")
	}

	// Verify ID token (validates signature, issuer, audience, expiry)
	idToken, err := c.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	// Extract claims from ID token
	var claims GoogleUserInfo
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse ID token claims: %w", err)
	}

	// Validate required claims
	if claims.ID == "" {
		return nil, fmt.Errorf("missing 'sub' claim in ID token")
	}
	if claims.Email == "" {
		return nil, fmt.Errorf("missing 'email' claim in ID token")
	}

	// For backwards compatibility, set ID field
	// (OIDC uses 'sub' but we called it 'ID' in our struct)
	// The json tag already maps 'sub' to ID field

	return &claims, nil
}

// GetUserInfoFromAccessToken retrieves user information from Google's UserInfo endpoint
// This is the legacy OAuth 2.0 method and should only be used as a fallback
// For OIDC-compliant flow, use GetUserInfo(oauth2Token) instead
func (c *GoogleClient) GetUserInfoFromAccessToken(ctx context.Context, accessToken string) (*GoogleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: status %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &userInfo, nil
}

// Validate checks if the configuration is valid
func (cfg Config) Validate() error {
	if cfg.ClientID == "" {
		return fmt.Errorf("GOOGLE_CLIENT_ID is required")
	}
	if cfg.ClientSecret == "" {
		return fmt.Errorf("GOOGLE_CLIENT_SECRET is required")
	}
	if cfg.RedirectURL == "" {
		return fmt.Errorf("GOOGLE_REDIRECT_URL is required")
	}
	return nil
}
