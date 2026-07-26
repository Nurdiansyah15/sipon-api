package web_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"sipon-api/internal/testhelper"
)

var (
	testSrv           *testhelper.TestServer
	testSuperadminID  string
	testSuperadminTok string
)

func TestMain(m *testing.M) {
	var cleanup func()
	testSrv, cleanup = testhelper.MustStartTestServer()

	// Buat user superadmin yang dipakai oleh role-permission tests
	testSuperadminID, testSuperadminTok = mustSetupSuperadmin()

	code := m.Run()
	cleanup()
	os.Exit(code)
}

// mustSetupSuperadmin mendaftarkan user baru, lalu langsung menyisipkan user_role superadmin.
func mustSetupSuperadmin() (userID, token string) {
	email := fmt.Sprintf("superadmin+%s@test.local", uuid.New().String()[:8])
	username := fmt.Sprintf("sadmin_%s", uuid.New().String()[:8])
	password := "Secret123!"

	// Register
	w := testSrv.POST("/api/v1/web/auth/register", map[string]any{
		"username": username,
		"email":    email,
		"password": password,
	})
	if w.Code != 201 {
		panic(fmt.Sprintf("mustSetupSuperadmin: register gagal %d — %s", w.Code, w.Body.String()))
	}

	// Login untuk dapat token
	w = testSrv.POST("/api/v1/web/auth/login", map[string]any{
		"identifier": email,
		"password":   password,
	})
	if w.Code != 200 {
		panic(fmt.Sprintf("mustSetupSuperadmin: login gagal %d — %s", w.Code, w.Body.String()))
	}

	var loginResp struct {
		Data struct {
			Token string `json:"token"`
			User  struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := parseJSON(w.Body.Bytes(), &loginResp); err != nil {
		panic("mustSetupSuperadmin: parse login response: " + err.Error())
	}
	uid := loginResp.Data.User.ID
	tok := loginResp.Data.Token

	// Cari ID superadmin role lalu assign langsung via DB
	ctx := context.Background()
	var superadminRoleID string
	if err := testSrv.DB.QueryRowContext(ctx,
		`SELECT id FROM roles WHERE name = 'superadmin' LIMIT 1`,
	).Scan(&superadminRoleID); err != nil {
		panic("mustSetupSuperadmin: cari superadmin role: " + err.Error())
	}

	testhelper.MustAssignUserRole(ctx, nil, testSrv.DB, uid, superadminRoleID, uid)

	// Tunggu sebentar agar principal cache tidak cache data lama (no-op cache, tapi wait tetap aman)
	_ = time.Millisecond

	return uid, tok
}
