package web_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"sipon-api/internal/testhelper"
)

// ── Logout ───────────────────────────────────────────────────────────────────
// Endpoint universal (bukan /web) — lihat router.Setup.

func TestAuthLogout_Unauthenticated(t *testing.T) {
	w := testSrv.POST("/api/v1/auth/logout", nil)
	assert.Equal(t, 401, w.Code)
}

func TestAuthLogout_Success_RevokesCurrentSession(t *testing.T) {
	email := uniqueEmail("logoutok")
	_ = mustRegisterUser(t, email, uniqueUsername("lgo"), "Secret123!")
	token, _ := mustLogin(t, email, "Secret123!")

	// Token masih valid sebelum logout.
	w := testSrv.GET("/api/v1/auth/session", testhelper.BearerHeader(token))
	assert.Equal(t, 200, w.Code, "session sebelum logout: %s", w.Body.String())

	w = testSrv.POST("/api/v1/auth/logout", nil, testhelper.BearerHeader(token))
	assert.Equal(t, 200, w.Code, "logout: %s", w.Body.String())

	// Access token yang sama sekarang harus ditolak.
	w = testSrv.GET("/api/v1/auth/session", testhelper.BearerHeader(token))
	assert.Equal(t, 401, w.Code, "session setelah logout harus 401: %s", w.Body.String())
}
