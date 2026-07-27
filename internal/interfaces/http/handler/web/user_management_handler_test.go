package web_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sipon-api/internal/testhelper"
)

// ── Helpers ───────────────────────────────────────────────────────────────────

// mustSetupUserWithRole mendaftarkan user baru (member default), lalu menyisipkan
// user_role untuk role tertentu via DB langsung. Mengembalikan (userID, token).
func mustSetupUserWithRole(t *testing.T, roleName string) (string, string) {
	t.Helper()
	email := uniqueEmail(roleName)
	username := uniqueUsername(roleName)
	_ = mustRegisterUser(t, email, username, "Secret123!")
	token, _ := mustLogin(t, email, "Secret123!")

	var resp struct {
		Data struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	w := testSrv.POST("/api/v1/web/auth/login", map[string]any{
		"identifier": email,
		"password":   "Secret123!",
	})
	require.Equal(t, 200, w.Code, "login: %s", w.Body.String())
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	uid := resp.Data.User.ID

	ctx := context.Background()
	var roleID string
	err := testSrv.DB.QueryRowContext(ctx, `SELECT id FROM roles WHERE name = $1 LIMIT 1`, roleName).Scan(&roleID)
	require.NoError(t, err, "cari role %s", roleName)
	testhelper.MustAssignUserRole(ctx, nil, testSrv.DB, uid, roleID, uid)

	return uid, token
}

// parseCreateUserResp membaca response CreateUser payload ke struct ringkas.
type createUserResp struct {
	ID                string `json:"id"`
	GeneratedPassword string `json:"generated_password"`
	Status            string `json:"status"`
}

// ── Unauthenticated (401) ────────────────────────────────────────────────────

func TestWebUserManagement_ListUsers_Unauthenticated(t *testing.T) {
	w := testSrv.GET("/api/v1/web/users")
	assert.Equal(t, 401, w.Code)
}

func TestWebUserManagement_GetUser_Unauthenticated(t *testing.T) {
	w := testSrv.GET("/api/v1/web/users/00000000-0000-0000-0000-000000000000")
	assert.Equal(t, 401, w.Code)
}

func TestWebUserManagement_CreateUser_Unauthenticated(t *testing.T) {
	w := testSrv.POST("/api/v1/web/users", map[string]any{
		"username": "should_not_create", "email": "x@example.com",
	})
	assert.Equal(t, 401, w.Code)
}

func TestWebUserManagement_ResetPassword_Unauthenticated(t *testing.T) {
	w := testSrv.POST("/api/v1/web/users/00000000-0000-0000-0000-000000000000/reset-password", nil)
	assert.Equal(t, 401, w.Code)
}

func TestWebUserManagement_Deactivate_Unauthenticated(t *testing.T) {
	w := testSrv.POST("/api/v1/web/users/00000000-0000-0000-0000-000000000000/deactivate", nil)
	assert.Equal(t, 401, w.Code)
}

// ── Forbidden (403) — member tidak punya permission admin ─────────────────────

func TestWebUserManagement_AsMember_Forbidden(t *testing.T) {
	email := uniqueEmail("mem")
	_ = mustRegisterUser(t, email, uniqueUsername("mem"), "Secret123!")
	token, _ := mustLogin(t, email, "Secret123!")
	h := testhelper.BearerHeader(token)

	t.Run("list", func(t *testing.T) {
		assert.Equal(t, 403, testSrv.GET("/api/v1/web/users", h).Code)
	})
	t.Run("get", func(t *testing.T) {
		assert.Equal(t, 403, testSrv.GET("/api/v1/web/users/00000000-0000-0000-0000-000000000000", h).Code)
	})
	t.Run("create", func(t *testing.T) {
		assert.Equal(t, 403, testSrv.POST("/api/v1/web/users", map[string]any{
			"username": "b", "email": "b@example.com",
		}, h).Code)
	})
	t.Run("reset-password", func(t *testing.T) {
		assert.Equal(t, 403, testSrv.POST("/api/v1/web/users/00000000-0000-0000-0000-000000000000/reset-password", nil, h).Code)
	})
	t.Run("deactivate", func(t *testing.T) {
		assert.Equal(t, 403, testSrv.POST("/api/v1/web/users/00000000-0000-0000-0000-000000000000/deactivate", nil, h).Code)
	})
	t.Run("reactivate", func(t *testing.T) {
		assert.Equal(t, 403, testSrv.POST("/api/v1/web/users/00000000-0000-0000-0000-000000000000/reactivate", nil, h).Code)
	})
}

// ── Validation (400/422) ─────────────────────────────────────────────────────

func TestWebUserManagement_CreateUser_MissingFields(t *testing.T) {
	w := testSrv.POST("/api/v1/web/users", map[string]any{
		// username & email sengaja dihilangkan (required)
	}, superadminHeader())
	assert.True(t, w.Code == 400 || w.Code == 422, "expected 400 or 422, got %d: %s", w.Code, w.Body.String())
}

func TestWebUserManagement_CreateUser_InvalidEmail(t *testing.T) {
	w := testSrv.POST("/api/v1/web/users", map[string]any{
		"username": uniqueUsername("iu"), "email": "not-an-email",
	}, superadminHeader())
	assert.True(t, w.Code == 400 || w.Code == 422, "expected 400 or 422, got %d: %s", w.Code, w.Body.String())
}

// ── Not Found (404) ──────────────────────────────────────────────────────────

func TestWebUserManagement_GetUser_NotFound(t *testing.T) {
	w := testSrv.GET("/api/v1/web/users/00000000-0000-0000-0000-000000000000", superadminHeader())
	assert.Equal(t, 404, w.Code)
}

func TestWebUserManagement_ResetPassword_NotFound(t *testing.T) {
	w := testSrv.POST("/api/v1/web/users/00000000-0000-0000-0000-000000000000/reset-password", nil, superadminHeader())
	assert.Equal(t, 404, w.Code)
}

func TestWebUserManagement_Deactivate_NotFound(t *testing.T) {
	w := testSrv.POST("/api/v1/web/users/00000000-0000-0000-0000-000000000000/deactivate", nil, superadminHeader())
	assert.Equal(t, 404, w.Code)
}

func TestWebUserManagement_Reactivate_NotFound(t *testing.T) {
	w := testSrv.POST("/api/v1/web/users/00000000-0000-0000-0000-000000000000/reactivate", nil, superadminHeader())
	assert.Equal(t, 404, w.Code)
}

// ── Success path (2xx) ───────────────────────────────────────────────────────
//
// Skenario penuh sebagai superadmin: create → get (verify roles mounted) →
// reset-password (verify new password shown & different) → deactivate → reactiv.
// Menggunakan endpoint kembali untuk memvalidasi behavior end-to-end.

func TestWebUserManagement_FullHappyPath_AsSuperadmin(t *testing.T) {
	newEmail := uniqueEmail("umcreate")
	newUsername := uniqueUsername("umcreate")

	// 1. Create
	w := testSrv.POST("/api/v1/web/users", map[string]any{
		"username": newUsername,
		"email":    newEmail,
		"fullname": "User Management Test",
	}, superadminHeader())
	require.Equal(t, 201, w.Code, "create: %s", w.Body.String())

	var created struct {
		Data createUserResp `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&created))
	require.NotEmpty(t, created.Data.ID)
	require.NotEmpty(t, created.Data.GeneratedPassword, "generated_password harus diisi (one-time reveal)")
	require.Equal(t, "ACTIVE", created.Data.Status)
	userID := created.Data.ID
	firstPw := created.Data.GeneratedPassword

	// 2. Get by id — roles diisi (new user belum punya assignment custom → kosong
	// tetapi field struct harus ke-serial; di sini cukup memastikan 200).
	w = testSrv.GET("/api/v1/web/users/"+userID, superadminHeader())
	require.Equal(t, 200, w.Code, "get: %s", w.Body.String())
	var got struct {
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Roles  []any  `json:"roles"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	assert.Equal(t, userID, got.Data.ID)
	assert.Equal(t, "ACTIVE", got.Data.Status)

	// 3. New user bisa login dengan password yang di-generate — bukti secret_hash disimpan.
	w = testSrv.POST("/api/v1/web/auth/login", map[string]any{
		"identifier": newEmail, "password": firstPw,
	})
	require.Equal(t, 200, w.Code, "login dengan generated password: %s", w.Body.String())

	// 4. Reset-password → generated_password baru (berbeda).
	w = testSrv.POST("/api/v1/web/users/"+userID+"/reset-password", nil, superadminHeader())
	require.Equal(t, 200, w.Code, "reset-password: %s", w.Body.String())
	var reset struct {
		Data struct {
			GeneratedPassword string `json:"generated_password"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&reset))
	require.NotEmpty(t, reset.Data.GeneratedPassword)
	require.NotEqual(t, firstPw, reset.Data.GeneratedPassword, "password baru harus berbeda")

	// 5. New user bisa login dengan password baru (reset benar-benar ter-apply).
	w = testSrv.POST("/api/v1/web/auth/login", map[string]any{
		"identifier": newEmail, "password": reset.Data.GeneratedPassword,
	})
	require.Equal(t, 200, w.Code, "login setelah reset: %s", w.Body.String())

	// 6. old password harus ditolak setelah reset.
	w = testSrv.POST("/api/v1/web/auth/login", map[string]any{
		"identifier": newEmail, "password": firstPw,
	})
	assert.NotEqual(t, 200, w.Code, "password lama seharusnya tidak berlaku lagi")

	// 7. Deactivate → status BANNED.
	w = testSrv.POST("/api/v1/web/users/"+userID+"/deactivate", nil, superadminHeader())
	require.Equal(t, 200, w.Code, "deactivate: %s", w.Body.String())
	var deactv struct {
		Data struct{ Status string `json:"status"` } `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&deactv))
	assert.Equal(t, "BANNED", deactv.Data.Status)

	// 8. Double-deactivate → 409 CodeUserAlreadyBanned.
	w = testSrv.POST("/api/v1/web/users/"+userID+"/deactivate", nil, superadminHeader())
	assert.Equal(t, 409, w.Code, "double-deactivate: %s", w.Body.String())

	// Banned user tidak bisa login.
	w = testSrv.POST("/api/v1/web/auth/login", map[string]any{
		"identifier": newEmail, "password": reset.Data.GeneratedPassword,
	})
	assert.NotEqual(t, 200, w.Code, "banned user tidak boleh login")

	// 9. Reactivate → status ACTIVE.
	w = testSrv.POST("/api/v1/web/users/"+userID+"/reactivate", nil, superadminHeader())
	require.Equal(t, 200, w.Code, "reactivate: %s", w.Body.String())
	var reactv struct {
		Data struct{ Status string `json:"status"` } `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&reactv))
	assert.Equal(t, "ACTIVE", reactv.Data.Status)

	// Double-reactivate → 409 CodeUserAlreadyActive.
	w = testSrv.POST("/api/v1/web/users/"+userID+"/reactivate", nil, superadminHeader())
	assert.Equal(t, 409, w.Code, "double-reactivate: %s", w.Body.String())
}

// ── List users (filtering) ───────────────────────────────────────────────────

func TestWebUserManagement_ListUsers_SuccessWithMeta(t *testing.T) {
	// Buat satu user baru agar count > 0 dapat dipaginate.
	newEmail := uniqueEmail("listusr")
	_ = mustRegisterUser(t, newEmail, uniqueUsername("listusr"), "Secret123!")

	w := testSrv.GET("/api/v1/web/users?limit=5&page=1", superadminHeader())
	require.Equal(t, 200, w.Code, w.Body.String())

	var resp struct {
		Data []struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		} `json:"data"`
		Meta struct {
			CurrentPage int64 `json:"current_page"`
			PerPage     int64 `json:"per_page"`
			Total       int64 `json:"total"`
		} `json:"meta"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.GreaterOrEqual(t, int(resp.Meta.Total), 1)
	assert.Equal(t, int64(5), resp.Meta.PerPage)
	assert.Equal(t, int64(1), resp.Meta.CurrentPage)
}

// ── Duplicate email/username (409) ───────────────────────────────────────────

func TestWebUserManagement_CreateUser_DuplicateEmail(t *testing.T) {
	// Register sekali lewat endpoint biasa dahulu untuk mengisi identity.
	existingEmail := uniqueEmail("dup")
	_ = mustRegisterUser(t, existingEmail, uniqueUsername("dup"), "Secret123!")

	w := testSrv.POST("/api/v1/web/users", map[string]any{
		"username": uniqueUsername("dup2"), "email": existingEmail,
	}, superadminHeader())
	assert.Equal(t, 409, w.Code, "duplicate email: %s", w.Body.String())
}

// ── Regression: admin role can reach manage_users-guarded routes AND
// assign_role-guarded user-roles routes (lockout bug fix), but cannot reach
// manage_roles-guarded POST /roles (admin lacks role definition rights). ────

func TestWebUserManagement_Admin_PermissionRegression(t *testing.T) {
	_, adminToken := mustSetupUserWithRole(t, "admin")
	adminH := testhelper.BearerHeader(adminToken)

	// manage_users-guarded: should 200/201 succeed for admin.
	t.Run("list users — admin can", func(t *testing.T) {
		assert.Equal(t, 200, testSrv.GET("/api/v1/web/users", adminH).Code)
	})

	// assign_role-guarded: POST /user-roles — admin can read route. We need a
	// valid target user_id & role_id to test the route actually reachable, but
	// guard passes before validation, so 4xx payload errors are also
	// acceptable here (they prove the route wasn't blocked by guard at 403).
	t.Run("assign user role — admin reaches route (not 403)", func(t *testing.T) {
		targetID := mustRegisterUser(t, uniqueEmail("admTarget"), uniqueUsername("admTarget"), "Secret123!")
		var memberRoleID string
		err := testSrv.DB.QueryRowContext(t.Context(), `SELECT id FROM roles WHERE name='member' LIMIT 1`).Scan(&memberRoleID)
		require.NoError(t, err)
		w := testSrv.POST("/api/v1/web/role-permission/user-roles", map[string]any{
			"user_id": targetID, "role_id": memberRoleID, "scope_type": "global",
		}, adminH)
		assert.Contains(t, []int{201, 409}, w.Code, "admin harus bisa assign-role (bug fix), got: %s", w.Body.String())
	})

	// GET /roles read-group — admin has assign_role, so read guard grants access.
	t.Run("list roles — admin can read (read group)", func(t *testing.T) {
		assert.Equal(t, 200, testSrv.GET("/api/v1/web/role-permission/roles", adminH).Code)
	})

	// POST /roles — manage_roles-guarded; admin lacks manage_roles → 403.
	t.Run("create role — admin forbidden (manage_roles missing)", func(t *testing.T) {
		w := testSrv.POST("/api/v1/web/role-permission/roles", map[string]any{
			"name": "regression_admin", "display_name": "Regression Admin Role",
			"role_type": "custom", "scope_type": "global",
		}, adminH)
		assert.Equal(t, 403, w.Code, "admin harus 403 pada manage_roles, got: %s", w.Body.String())
	})

	// manage_role_permissions-guarded: POST /roles/:id/permissions must 403 for admin.
	t.Run("assign role permission — admin forbidden (manage_role_permissions missing)", func(t *testing.T) {
		var memberRoleID string
		err := testSrv.DB.QueryRowContext(t.Context(), `SELECT id FROM roles WHERE name='member' LIMIT 1`).Scan(&memberRoleID)
		require.NoError(t, err)
		w := testSrv.POST("/api/v1/web/role-permission/roles/"+memberRoleID+"/permissions",
			map[string]any{"permission_key": "manage_users"}, adminH)
		// member is system role — would 409 if reached, but admin lacks the
		// permission so the guard blocks at 403 before the usecase runs.
		assert.Equal(t, 403, w.Code, "admin harus 403 pada manage_role_permissions, got: %s", w.Body.String())
	})
}