package jwt

import (
	"errors"
	"fmt"
	"sipon-api/internal/app/port"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTTokenGenerator struct {
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewJWTTokenGenerator(secretKey string, accessTTL, refreshTTL time.Duration) *JWTTokenGenerator {
	return &JWTTokenGenerator{
		secretKey:       []byte(secretKey),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

// accessClaims carries only identity — authz data is loaded from Redis/DB per-request.
type accessClaims struct {
	SessionID string `json:"sid"`
	DeviceID  string `json:"did,omitempty"`
	jwt.RegisteredClaims
}

type refreshClaims struct {
	UserID   string `json:"user_id"`
	DeviceID string `json:"did,omitempty"`
	jwt.RegisteredClaims
}

func (g *JWTTokenGenerator) GenerateAccessToken(userID, sessionID, deviceID string) (string, error) {
	claims := accessClaims{
		SessionID: sessionID,
		DeviceID:  deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(g.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(g.secretKey)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

func (g *JWTTokenGenerator) GenerateRefreshToken(userID, deviceID string) (string, error) {
	claims := refreshClaims{
		UserID:   userID,
		DeviceID: deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(g.refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(g.secretKey)
	if err != nil {
		return "", fmt.Errorf("sign refresh token: %w", err)
	}
	return signed, nil
}

func (g *JWTTokenGenerator) ParseAccessToken(tokenStr string) (*port.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &accessClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return g.secretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("token tidak valid: %w", err)
	}
	claims, ok := token.Claims.(*accessClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token claims tidak valid")
	}
	userID := claims.Subject
	if userID == "" {
		return nil, errors.New("token subject kosong")
	}
	var issuedAt time.Time
	if claims.IssuedAt != nil {
		issuedAt = claims.IssuedAt.Time
	}
	return &port.TokenClaims{
		UserID:    userID,
		SessionID: claims.SessionID,
		DeviceID:  claims.DeviceID,
		IssuedAt:  issuedAt,
	}, nil
}

func (g *JWTTokenGenerator) ParseRefreshToken(tokenStr string) (*port.RefreshTokenClaims, error) {
	if strings.TrimSpace(tokenStr) == "" {
		return nil, errors.New("refresh token kosong")
	}
	token, err := jwt.ParseWithClaims(tokenStr, &refreshClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return g.secretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("refresh token tidak valid: %w", err)
	}
	claims, ok := token.Claims.(*refreshClaims)
	if !ok || !token.Valid {
		return nil, errors.New("refresh token claims tidak valid")
	}
	if claims.UserID == "" {
		return nil, errors.New("refresh token user id kosong")
	}
	if claims.Subject == "" || claims.Subject != claims.UserID {
		return nil, errors.New("refresh token subject tidak valid")
	}
	var issuedAt time.Time
	if claims.IssuedAt != nil {
		issuedAt = claims.IssuedAt.Time
	}
	return &port.RefreshTokenClaims{UserID: claims.UserID, DeviceID: claims.DeviceID, IssuedAt: issuedAt}, nil
}
