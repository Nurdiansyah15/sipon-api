package web_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sipon-api/internal/testhelper"
)

// ── Register ─────────────────────────────────────────────────────────────────

func TestWebAuth_Register_Success(t *testing.T) {
	w := testSrv.POST("/api/v1/web/auth/register", map[string]any{
		"username": uniqueUsername("usr"),
		"email":    uniqueEmail("reg"),
		"password": "Secret123!",
	})
	assert.Equal(t, 201, w.Code)

	var resp struct {
		Data struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.NotEmpty(t, resp.Data.UserID)
}

func TestWebAuth_Register_DuplicateEmail(t *testing.T) {
	email := uniqueEmail("dup")
	_ = mustRegisterUser(t, email, uniqueUsername("d1"), "Secret123!")

	w := testSrv.POST("/api/v1/web/auth/register", map[string]any{
		"username": uniqueUsername("d2"),
		"email":    email,
		"password": "Secret123!",
	})
	assert.Equal(t, 409, w.Code)
}

func TestWebAuth_Register_DuplicateUsername(t *testing.T) {
	username := uniqueUsername("dupusr")
	_ = mustRegisterUser(t, uniqueEmail("e1"), username, "Secret123!")

	w := testSrv.POST("/api/v1/web/auth/register", map[string]any{
		"username": username,
		"email":    uniqueEmail("e2"),
		"password": "Secret123!",
	})
	assert.Equal(t, 409, w.Code)
}

func TestWebAuth_Register_MissingFields(t *testing.T) {
	cases := []map[string]any{
		{"email": uniqueEmail("miss"), "password": "Secret123!"},                             // username missing
		{"username": uniqueUsername("m"), "password": "Secret123!"},                          // email missing
		{"username": uniqueUsername("m"), "email": uniqueEmail("m")},                         // password missing
		{"username": "ab", "email": uniqueEmail("short"), "password": "Secret123!"},          // username too short
		{"username": uniqueUsername("m"), "email": "not-an-email", "password": "Secret123!"}, // invalid email
	}
	for _, body := range cases {
		w := testSrv.POST("/api/v1/web/auth/register", body)
		assert.Equal(t, 422, w.Code, "body=%v response=%s", body, w.Body.String())
	}
}

// ── Login ─────────────────────────────────────────────────────────────────────

func TestWebAuth_Login_Success(t *testing.T) {
	email := uniqueEmail("login")
	password := "Secret123!"
	_ = mustRegisterUser(t, email, uniqueUsername("login"), password)

	w := testSrv.POST("/api/v1/web/auth/login", map[string]any{
		"identifier": email,
		"password":   password,
	})
	assert.Equal(t, 200, w.Code)

	var resp struct {
		Data struct {
			Token        string `json:"token"`
			RefreshToken string `json:"refresh_token"`
			User         struct {
				Email string `json:"email"`
			} `json:"user"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.NotEmpty(t, resp.Data.Token)
	assert.NotEmpty(t, resp.Data.RefreshToken)
	assert.Equal(t, email, resp.Data.User.Email)
}

func TestWebAuth_Login_WrongPassword(t *testing.T) {
	email := uniqueEmail("badpw")
	_ = mustRegisterUser(t, email, uniqueUsername("badpw"), "Secret123!")

	w := testSrv.POST("/api/v1/web/auth/login", map[string]any{
		"identifier": email,
		"password":   "WrongPassword99!",
	})
	assert.Equal(t, 401, w.Code)
}

func TestWebAuth_Login_UserNotFound(t *testing.T) {
	w := testSrv.POST("/api/v1/web/auth/login", map[string]any{
		"identifier": "nonexistent@test.local",
		"password":   "Secret123!",
	})
	assert.Equal(t, 401, w.Code)
}

func TestWebAuth_Login_MissingFields(t *testing.T) {
	w := testSrv.POST("/api/v1/web/auth/login", map[string]any{
		"identifier": "only-identifier@test.local",
	})
	assert.Equal(t, 422, w.Code)
}

func TestWebAuth_Login_ByUsername(t *testing.T) {
	username := uniqueUsername("byusr")
	password := "Secret123!"
	_ = mustRegisterUser(t, uniqueEmail("byusr"), username, password)

	w := testSrv.POST("/api/v1/web/auth/login", map[string]any{
		"identifier": username,
		"password":   password,
	})
	assert.Equal(t, 200, w.Code)
}

func TestWebAuth_Login_AccountLockout(t *testing.T) {
	email := uniqueEmail("lkout")
	_ = mustRegisterUser(t, email, uniqueUsername("lkout"), "Secret123!")

	// 5 wrong attempts → 401 each, lockout triggered after 5th
	for i := 0; i < 5; i++ {
		w := testSrv.POST("/api/v1/web/auth/login", map[string]any{
			"identifier": email,
			"password":   "WrongPassword99!",
		})
	assert.Equal(t, 422, w.Code)
}

	// 6th attempt: account is now locked → 429
	w := testSrv.POST("/api/v1/web/auth/login", map[string]any{
		"identifier": email,
		"password":   "WrongPassword99!",
	})
	assert.Equal(t, 429, w.Code)

	// Even correct password is rejected while locked
	w = testSrv.POST("/api/v1/web/auth/login", map[string]any{
		"identifier": email,
		"password":   "Secret123!",
	})
	assert.Equal(t, 429, w.Code)
}

func TestWebAuth_Login_ResetPasswordUnlocksAccount(t *testing.T) {
	email := uniqueEmail("unlock")
	userID := mustRegisterUser(t, email, uniqueUsername("unlock"), "Secret123!")

	// Lock the account via 5 wrong attempts
	for i := 0; i < 5; i++ {
		testSrv.POST("/api/v1/web/auth/login", map[string]any{
			"identifier": email,
			"password":   "WrongPassword99!",
		})
	}

	// Verify account is locked
	w := testSrv.POST("/api/v1/web/auth/login", map[string]any{
		"identifier": email,
		"password":   "Secret123!",
	})
	assert.Equal(t, 429, w.Code)

	// Mark email as verified directly in DB so forgot-password OTP can be sent
	_, err := testSrv.DB.ExecContext(t.Context(), `
		UPDATE user_identities SET status='VERIFIED', verified_at=NOW()
		WHERE user_id=$1 AND kind='EMAIL'`, userID)
	require.NoError(t, err)

	w = testSrv.POST("/api/v1/web/auth/password/forgot", map[string]any{
		"email": email,
	})
	assert.Equal(t, 200, w.Code)

	otp := mustGetOTPFromDB(t, userID, "RESET_PASSWORD")

	w = testSrv.POST("/api/v1/web/auth/password/reset", map[string]any{
		"email":    email,
		"token":    otp,
		"password": "NewSecret456!",
	})
	assert.Equal(t, 200, w.Code)

	// Account should be unlocked — login with new password succeeds
	w = testSrv.POST("/api/v1/web/auth/login", map[string]any{
		"identifier": email,
		"password":   "NewSecret456!",
	})
	assert.Equal(t, 200, w.Code)
}

// ── Me ────────────────────────────────────────────────────────────────────────

func TestWebAuth_Me_Authenticated(t *testing.T) {
	email := uniqueEmail("metest")
	_ = mustRegisterUser(t, email, uniqueUsername("me"), "Secret123!")
	token, _ := mustLogin(t, email, "Secret123!")

	w := testSrv.GET("/api/v1/web/auth/me", testhelper.BearerHeader(token))
	assert.Equal(t, 200, w.Code)

	var resp struct {
		Data struct {
			Email string `json:"email"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, email, resp.Data.Email)
}

func TestWebAuth_Me_Unauthenticated(t *testing.T) {
	w := testSrv.GET("/api/v1/web/auth/me")
	assert.Equal(t, 401, w.Code)
}

func TestWebAuth_Me_InvalidToken(t *testing.T) {
	w := testSrv.GET("/api/v1/web/auth/me", testhelper.BearerHeader("invalid.jwt.token"))
	assert.Equal(t, 401, w.Code)
}

// ── Refresh Token ──────────────────────────────────────────────────────────────

func TestWebAuth_RefreshToken_Success(t *testing.T) {
	email := uniqueEmail("refresh")
	_ = mustRegisterUser(t, email, uniqueUsername("ref"), "Secret123!")
	_, refreshToken := mustLogin(t, email, "Secret123!")

	w := testSrv.POST("/api/v1/web/auth/refresh-token", map[string]any{
		"refresh_token": refreshToken,
	})
	assert.Equal(t, 200, w.Code)

	var resp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.NotEmpty(t, resp.Data.Token)
}

func TestWebAuth_RefreshToken_Invalid(t *testing.T) {
	w := testSrv.POST("/api/v1/web/auth/refresh-token", map[string]any{
		"refresh_token": "invalid-refresh-token",
	})
	assert.Equal(t, 401, w.Code)
}

// ── Forgot Password ───────────────────────────────────────────────────────────

func TestWebAuth_ForgotPassword_ExistingEmail(t *testing.T) {
	email := uniqueEmail("forgot")
	userID := mustRegisterUser(t, email, uniqueUsername("fgt"), "Secret123!")
	// ForgotPassword hanya bekerja untuk email yang sudah diverifikasi
	testhelper.MustVerifyUserEmailIdentity(t.Context(), t, testSrv.DB, userID)

	w := testSrv.POST("/api/v1/web/auth/password/forgot", map[string]any{
		"email": email,
	})
	assert.Equal(t, 200, w.Code)
}

func TestWebAuth_ForgotPassword_NonExistingEmail(t *testing.T) {
	w := testSrv.POST("/api/v1/web/auth/password/forgot", map[string]any{
		"email": "nonexistent+forgot@test.local",
	})
	// API sengaja mengembalikan 200 untuk email yang tidak terdaftar (security: avoid email enumeration)
	assert.Equal(t, 200, w.Code)
}

func TestWebAuth_ForgotPassword_InvalidEmail(t *testing.T) {
	w := testSrv.POST("/api/v1/web/auth/password/forgot", map[string]any{
		"email": "not-an-email",
	})
	assert.Equal(t, 422, w.Code)
}

// ── Change Password ───────────────────────────────────────────────────────────

func TestWebAuth_ChangePassword_Success(t *testing.T) {
	email := uniqueEmail("chpw")
	password := "Secret123!"
	_ = mustRegisterUser(t, email, uniqueUsername("chpw"), password)
	token, _ := mustLogin(t, email, password)

	w := testSrv.POST("/api/v1/web/auth/change-password",
		map[string]any{
			"current_password": password,
			"new_password":     "NewSecret456!",
		},
		testhelper.BearerHeader(token),
	)
	assert.Equal(t, 200, w.Code)
}

func TestWebAuth_ChangePassword_WrongCurrent(t *testing.T) {
	email := uniqueEmail("chpwbad")
	_ = mustRegisterUser(t, email, uniqueUsername("cpbad"), "Secret123!")
	token, _ := mustLogin(t, email, "Secret123!")

	w := testSrv.POST("/api/v1/web/auth/change-password",
		map[string]any{
			"current_password": "WrongOld123!",
			"new_password":     "NewSecret456!",
		},
		testhelper.BearerHeader(token),
	)
	assert.Equal(t, 401, w.Code)
}

func TestWebAuth_ChangePassword_Unauthenticated(t *testing.T) {
	w := testSrv.POST("/api/v1/web/auth/change-password", map[string]any{
		"current_password": "Secret123!",
		"new_password":     "NewSecret456!",
	})
	assert.Equal(t, 401, w.Code)
}

// ── Set Password (first-time, social-login-only account) ────────────────────

func TestWebAuth_SetPassword_Success(t *testing.T) {
	email := uniqueEmail("setpw")
	password := "Secret123!"
	userID := mustRegisterUser(t, email, uniqueUsername("setpw"), password)
	token, _ := mustLogin(t, email, password)

	// Simulasikan akun social-login-only: credential LOCAL ada tapi belum
	// punya password (persis kondisi user Google/Apple hasil createNewGoogleUser).
	_, err := testSrv.DB.ExecContext(t.Context(), `
		UPDATE credentials SET secret_hash=NULL WHERE user_id=$1 AND type='LOCAL'`, userID)
	require.NoError(t, err)

	w := testSrv.POST("/api/v1/web/auth/set-password",
		map[string]any{"new_password": "NewSecret456!"},
		testhelper.BearerHeader(token),
	)
	assert.Equal(t, 200, w.Code, w.Body.String())

	// Password baru harus langsung bisa dipakai login.
	w2 := testSrv.POST("/api/v1/web/auth/login", map[string]any{
		"identifier": email,
		"password":   "NewSecret456!",
	})
	assert.Equal(t, 200, w2.Code)
}

func TestWebAuth_SetPassword_AlreadyHasPassword(t *testing.T) {
	email := uniqueEmail("setpwexist")
	password := "Secret123!"
	_ = mustRegisterUser(t, email, uniqueUsername("setpwexist"), password)
	token, _ := mustLogin(t, email, password)

	w := testSrv.POST("/api/v1/web/auth/set-password",
		map[string]any{"new_password": "NewSecret456!"},
		testhelper.BearerHeader(token),
	)
	assert.Equal(t, 409, w.Code, w.Body.String())
}

func TestWebAuth_SetPassword_InvalidPassword(t *testing.T) {
	email := uniqueEmail("setpwinvalid")
	password := "Secret123!"
	userID := mustRegisterUser(t, email, uniqueUsername("setpwinvalid"), password)
	token, _ := mustLogin(t, email, password)

	_, err := testSrv.DB.ExecContext(t.Context(), `
		UPDATE credentials SET secret_hash=NULL WHERE user_id=$1 AND type='LOCAL'`, userID)
	require.NoError(t, err)

	w := testSrv.POST("/api/v1/web/auth/set-password",
		map[string]any{"new_password": "short"},
		testhelper.BearerHeader(token),
	)
	assert.Equal(t, 422, w.Code, w.Body.String())
}

func TestWebAuth_SetPassword_Unauthenticated(t *testing.T) {
	w := testSrv.POST("/api/v1/web/auth/set-password", map[string]any{
		"new_password": "NewSecret456!",
	})
	assert.Equal(t, 401, w.Code)
}
