package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Roles     []string  `json:"roles"`
	TokenType TokenType `json:"token_type"`
	SessionID string    `json:"session_id"`
	jwt.RegisteredClaims
}

type TokenDetails struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	AtExpires    int64  `json:"at_expires"`
	RtExpires    int64  `json:"rt_expires"`
	SessionID    string `json:"session_id"`
}

type JWTManager struct {
	accessSecret  string
	refreshSecret string
	issuer        string
	audience      string
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewJWTManager(accessSecret, refreshSecret, issuer, audience string, accessTTL, refreshTTL time.Duration) *JWTManager {
	return &JWTManager{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		issuer:        issuer,
		audience:      audience,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (m *JWTManager) GenerateTokens(userID uuid.UUID, email string, roles []string) (*TokenDetails, error) {
	sessionID := uuid.New().String()
	now := time.Now()
	atExpires := now.Add(m.accessTTL)
	rtExpires := now.Add(m.refreshTTL)

	// Generate Access Token
	accessToken, err := m.generateToken(userID, email, roles, AccessToken, sessionID, now, atExpires, m.accessSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate Refresh Token
	refreshToken, err := m.generateToken(userID, email, roles, RefreshToken, sessionID, now, rtExpires, m.refreshSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenDetails{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AtExpires:    atExpires.Unix(),
		RtExpires:    rtExpires.Unix(),
		SessionID:    sessionID,
	}, nil
}

func (m *JWTManager) generateToken(userID uuid.UUID, email string, roles []string, tokenType TokenType, sessionID string, issuedAt, expiresAt time.Time, secret string) (string, error) {
	claims := &Claims{
		UserID:    userID,
		Email:     email,
		Roles:     roles,
		TokenType: tokenType,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			NotBefore: jwt.NewNumericDate(issuedAt),
			Issuer:    m.issuer,
			Subject:   userID.String(),
			Audience:  jwt.ClaimStrings{m.audience},
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := m.validateToken(tokenString, m.accessSecret)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != AccessToken {
		return nil, errors.New("invalid token type")
	}
	return claims, nil
}

func (m *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := m.validateToken(tokenString, m.refreshSecret)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != RefreshToken {
		return nil, errors.New("invalid token type")
	}
	return claims, nil
}

func (m *JWTManager) validateToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(m.issuer), jwt.WithAudience(m.audience))

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, errors.New("token has expired")
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, errors.New("token is not valid yet")
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, errors.New("malformed token")
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, errors.New("invalid token signature")
		case errors.Is(err, jwt.ErrTokenInvalidIssuer):
			return nil, errors.New("invalid token issuer")
		case errors.Is(err, jwt.ErrTokenInvalidAudience):
			return nil, errors.New("invalid token audience")
		default:
			return nil, fmt.Errorf("invalid token: %w", err)
		}
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func (m *JWTManager) ValidateTokenType(tokenString string) (TokenType, error) {
	// Parse without verification to get token type
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims.TokenType, nil
	}

	return "", errors.New("invalid token claims")
}

// GenerateSecureSecret generates a cryptographically secure random secret
func GenerateSecureSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
