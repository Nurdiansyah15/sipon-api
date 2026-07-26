package web_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// parseJSON adalah alias json.Unmarshal yang dipakai di main_test.go.
func parseJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// mustRegisterUser mendaftarkan user baru dengan email/username unik dan mengembalikan userID.
func mustRegisterUser(t *testing.T, email, username, password string) string {
	t.Helper()
	w := testSrv.POST("/api/v1/web/auth/register", map[string]any{
		"username": username,
		"email":    email,
		"password": password,
	})
	require.Equal(t, 201, w.Code, "register: %s", w.Body.String())
	var resp struct {
		Data struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	return resp.Data.UserID
}

// mustLogin melakukan login dan mengembalikan (accessToken, refreshToken).
func mustLogin(t *testing.T, identifier, password string) (accessToken, refreshToken string) {
	t.Helper()
	w := testSrv.POST("/api/v1/web/auth/login", map[string]any{
		"identifier": identifier,
		"password":   password,
	})
	require.Equal(t, 200, w.Code, "login: %s", w.Body.String())
	var resp struct {
		Data struct {
			Token        string `json:"token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	return resp.Data.Token, resp.Data.RefreshToken
}

// uniqueEmail menghasilkan email unik untuk test isolation.
func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s+%s@test.local", prefix, uuid.New().String()[:8])
}

// uniqueUsername menghasilkan username unik untuk test isolation.
func uniqueUsername(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, uuid.New().String()[:8])
}

// mustGetOTPFromDB membaca kode OTP terbaru dari DB untuk user dan purpose tertentu.
// purpose: "EMAIL_VERIFICATION", "RESET_PASSWORD", "CHANGE_EMAIL", "CHANGE_PHONE"
func mustGetOTPFromDB(t *testing.T, userID, purpose string) string {
	t.Helper()
	var code string
	err := testSrv.DB.QueryRowContext(t.Context(),
		`SELECT code FROM verification_codes WHERE user_id=$1 AND purpose=$2 ORDER BY created_at DESC LIMIT 1`,
		userID, purpose,
	).Scan(&code)
	require.NoError(t, err, "mustGetOTPFromDB: OTP tidak ditemukan untuk user %s purpose %s", userID, purpose)
	return code
}
