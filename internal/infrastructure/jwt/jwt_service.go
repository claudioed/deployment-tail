package jwt

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/claudioed/deployment-tail/internal/domain/user"
)

// JWTService handles JWT token operations
type JWTService struct {
	secret    []byte
	expiryDur time.Duration
	issuer    string
}

// Claims represents the JWT claims
type Claims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Config holds JWT configuration
type Config struct {
	Secret string
	Expiry time.Duration
	Issuer string
}

// NewJWTService creates a new JWT service
func NewJWTService(cfg Config) (*JWTService, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &JWTService{
		secret:    []byte(cfg.Secret),
		expiryDur: cfg.Expiry,
		issuer:    cfg.Issuer,
	}, nil
}

// GenerateToken creates a signed JWT for a user
func (s *JWTService) GenerateToken(u *user.User) (string, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.expiryDur)

	claims := Claims{
		UserID: u.ID().String(),
		Email:  u.Email().String(),
		Role:   u.Role().String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Subject:   u.ID().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken verifies a token's signature and claims
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now().UTC()) {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}

// RefreshToken issues a new token with extended expiry
func (s *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("cannot refresh invalid token: %w", err)
	}

	// Create new token with extended expiry
	now := time.Now().UTC()
	expiresAt := now.Add(s.expiryDur)

	newClaims := Claims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Subject:   claims.UserID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign refreshed token: %w", err)
	}

	return signedToken, nil
}

// ParseClaims extracts claims from a token without full validation
func (s *JWTService) ParseClaims(tokenString string) (*Claims, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// HashToken creates a SHA256 hash of a token for storage
func (s *JWTService) HashToken(tokenString string) string {
	hash := sha256.Sum256([]byte(tokenString))
	return hex.EncodeToString(hash[:])
}

// Validate checks if the configuration is valid
func (cfg Config) Validate() error {
	if cfg.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters for security")
	}
	if cfg.Expiry <= 0 {
		return fmt.Errorf("JWT_EXPIRY must be positive")
	}
	if cfg.Issuer == "" {
		cfg.Issuer = "deployment-tail"
	}
	return nil
}
