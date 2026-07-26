package port

import "time"

type TokenClaims struct {
	UserID    string
	SessionID string
	// DeviceID adalah device_id client-generated, kosong kalau client tidak
	// mengirimnya saat login. Dipakai untuk revocation per-device — lihat
	// SessionRevocationStore.
	DeviceID string
	IssuedAt time.Time
}

// RefreshTokenClaims membawa IssuedAt & DeviceID supaya pemanggil
// (RefreshTokenUseCase) bisa mengecek terhadap SessionRevocationStore baik
// yang scope user (logout-all) maupun scope device (logout device lain).
type RefreshTokenClaims struct {
	UserID   string
	DeviceID string
	IssuedAt time.Time
}

// TokenGenerator abstraksi pembuatan JWT.
// Access token membawa sub (user_id), sid (session_id), dan opsional did
// (device_id). Data authz (roles, permissions) diload dari Redis/DB via
// PrincipalLoader — bukan dari token.
type TokenGenerator interface {
	GenerateAccessToken(userID, sessionID, deviceID string) (string, error)
	GenerateRefreshToken(userID, deviceID string) (string, error)
	ParseAccessToken(token string) (*TokenClaims, error)
	ParseRefreshToken(token string) (*RefreshTokenClaims, error)
}
