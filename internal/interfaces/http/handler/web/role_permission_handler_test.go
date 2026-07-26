package web_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sipon-api/internal/testhelper"
)

// superadminHeader mengembalikan header Authorization dengan token superadmin.
func superadminHeader() map[string]string {
	return testhelper.BearerHeader(testSuperadminTok)
}

// ── Roles ─────────────────────────────────────────────────────────────────────

func TestWebRolePerm_ListRoles_AsSuperadmin(t *testing.T) {
	w := testSrv.GET("/api/v1/web/role-permission/roles", superadminHeader())
	assert.Equal(t, 200, w.Code)
}

func TestWebRolePerm_ListRoles_AsRegularUser(t *testing.T) {
	email := uniqueEmail("norole")
	_ = mustRegisterUser(t, email, uniqueUsername("nr"), "Secret123!")
	token, _ := mustLogin(t, email, "Secret123!")

	w := testSrv.GET("/api/v1/web/role-permission/roles", testhelper.BearerHeader(token))
	assert.Equal(t, 403, w.Code)
}

func TestWebRolePerm_ListRoles_Unauthenticated(t *testing.T) {
	w := testSrv.GET("/api/v1/web/role-permission/roles")
	assert.Equal(t, 401, w.Code)
}

func TestWebRolePerm_CreateRole_Success(t *testing.T) {
	roleName := "test_role_" + uniqueUsername("role")
	w := testSrv.POST("/api/v1/web/role-permission/roles",
		map[string]any{
			"name":         roleName,
			"display_name": "Test Role",
			"role_type":    "custom",
			"scope_type":   "global",
		},
		superadminHeader(),
	)
	assert.Equal(t, 201, w.Code, w.Body.String())

	var resp struct {
		Data struct {
			ID          string   `json:"id"`
			Permissions []string `json:"permissions"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.NotEmpty(t, resp.Data.ID)
	// Custom role tidak ada di constant.RolePermissions → tidak punya permission apa pun.
	assert.Empty(t, resp.Data.Permissions)
}

func TestWebRolePerm_CreateRole_Forbidden(t *testing.T) {
	email := uniqueEmail("norole2")
	_ = mustRegisterUser(t, email, uniqueUsername("nr2"), "Secret123!")
	token, _ := mustLogin(t, email, "Secret123!")

	w := testSrv.POST("/api/v1/web/role-permission/roles",
		map[string]any{
			"name":         "forbidden_role",
			"display_name": "Forbidden Role",
			"role_type":    "custom",
			"scope_type":   "global",
		},
		testhelper.BearerHeader(token),
	)
	assert.Equal(t, 403, w.Code)
}

func TestWebRolePerm_CreateRole_InvalidPayload(t *testing.T) {
	// role_type harus salah satu dari "system"/"custom" — kirim nilai tak valid.
	w := testSrv.POST("/api/v1/web/role-permission/roles",
		map[string]any{
			"name":         "invalid_role_" + uniqueUsername("r"),
			"display_name": "Invalid Role",
			"role_type":    "not-a-valid-type",
			"scope_type":   "global",
		},
		superadminHeader(),
	)
	assert.True(t, w.Code == 400 || w.Code == 422, "expected 400 or 422, got %d: %s", w.Code, w.Body.String())
}

func TestWebRolePerm_GetRole_NotFound(t *testing.T) {
	w := testSrv.GET("/api/v1/web/role-permission/roles/00000000-0000-0000-0000-000000000000",
		superadminHeader(),
	)
	assert.Equal(t, 404, w.Code)
}

func TestWebRolePerm_GetRole_Success(t *testing.T) {
	var superadminRoleID string
	err := testSrv.DB.QueryRowContext(t.Context(), `SELECT id FROM roles WHERE name = 'superadmin' LIMIT 1`).Scan(&superadminRoleID)
	require.NoError(t, err)

	w := testSrv.GET("/api/v1/web/role-permission/roles/"+superadminRoleID, superadminHeader())
	assert.Equal(t, 200, w.Code)

	var resp struct {
		Data struct {
			Name        string   `json:"name"`
			Permissions []string `json:"permissions"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "superadmin", resp.Data.Name)
	// superadmin punya permission via constant.RolePermissions (lihat permission_constant.go).
	assert.Contains(t, resp.Data.Permissions, "manage_system_settings")
}

// ── User Roles ────────────────────────────────────────────────────────────────

func TestWebRolePerm_ListUserRoles_AsSuperadmin(t *testing.T) {
	w := testSrv.GET("/api/v1/web/role-permission/user-roles", superadminHeader())
	assert.Equal(t, 200, w.Code)
}

func TestWebRolePerm_ListUserRoles_Unauthenticated(t *testing.T) {
	w := testSrv.GET("/api/v1/web/role-permission/user-roles")
	assert.Equal(t, 401, w.Code)
}

func TestWebRolePerm_AssignUserRole_Success(t *testing.T) {
	// Buat target user
	targetEmail := uniqueEmail("targetusr")
	targetID := mustRegisterUser(t, targetEmail, uniqueUsername("tu"), "Secret123!")

	// Cari member role ID
	var memberRoleID string
	err := testSrv.DB.QueryRowContext(t.Context(), `SELECT id FROM roles WHERE name = 'member' LIMIT 1`).Scan(&memberRoleID)
	require.NoError(t, err, "get member role id")

	w := testSrv.POST("/api/v1/web/role-permission/user-roles",
		map[string]any{
			"user_id":    targetID,
			"role_id":    memberRoleID,
			"scope_type": "global",
		},
		superadminHeader(),
	)
	// 201 atau 409 (jika user sudah punya role member dari register)
	assert.Contains(t, []int{201, 409}, w.Code, "body=%s", w.Body.String())
}

func TestWebRolePerm_AssignUserRole_InvalidPayload(t *testing.T) {
	w := testSrv.POST("/api/v1/web/role-permission/user-roles",
		map[string]any{
			"user_id": "some-user-id",
			// role_id sengaja dihilangkan (required)
			"scope_type": "global",
		},
		superadminHeader(),
	)
	assert.True(t, w.Code == 400 || w.Code == 422, "expected 400 or 422, got %d: %s", w.Code, w.Body.String())
}

func TestWebRolePerm_AssignUserRole_Forbidden(t *testing.T) {
	email := uniqueEmail("noassign")
	_ = mustRegisterUser(t, email, uniqueUsername("na"), "Secret123!")
	token, _ := mustLogin(t, email, "Secret123!")

	w := testSrv.POST("/api/v1/web/role-permission/user-roles",
		map[string]any{
			"user_id":    "00000000-0000-0000-0000-000000000000",
			"role_id":    "00000000-0000-0000-0000-000000000000",
			"scope_type": "global",
		},
		testhelper.BearerHeader(token),
	)
	assert.Equal(t, 403, w.Code)
}

func TestWebRolePerm_GetUserRole_NotFound(t *testing.T) {
	w := testSrv.GET("/api/v1/web/role-permission/user-roles/00000000-0000-0000-0000-000000000000",
		superadminHeader(),
	)
	assert.Equal(t, 404, w.Code)
}

func TestWebRolePerm_DeactivateReactivateDeleteUserRole_Success(t *testing.T) {
	targetEmail := uniqueEmail("lifecycle")
	targetID := mustRegisterUser(t, targetEmail, uniqueUsername("lc"), "Secret123!")

	var adminRoleID string
	err := testSrv.DB.QueryRowContext(t.Context(), `SELECT id FROM roles WHERE name = 'admin' LIMIT 1`).Scan(&adminRoleID)
	require.NoError(t, err)

	w := testSrv.POST("/api/v1/web/role-permission/user-roles",
		map[string]any{
			"user_id":    targetID,
			"role_id":    adminRoleID,
			"scope_type": "global",
		},
		superadminHeader(),
	)
	require.Equal(t, 201, w.Code, w.Body.String())

	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&created))
	userRoleID := created.Data.ID
	require.NotEmpty(t, userRoleID)

	w = testSrv.POST("/api/v1/web/role-permission/user-roles/"+userRoleID+"/deactivate", nil, superadminHeader())
	assert.Equal(t, 200, w.Code, w.Body.String())

	w = testSrv.POST("/api/v1/web/role-permission/user-roles/"+userRoleID+"/reactivate", nil, superadminHeader())
	assert.Equal(t, 200, w.Code, w.Body.String())

	w = testSrv.DELETE("/api/v1/web/role-permission/user-roles/"+userRoleID, superadminHeader())
	assert.Equal(t, 200, w.Code, w.Body.String())
}

// ── Permission keys catalog & custom role permission assignment ─────────────
// Permission key & mapping role→permission role SYSTEM tetap fixed di kode
// (constant.RolePermissions) — hanya role CUSTOM yang permission-nya bisa
// diatur dinamis lewat endpoint di bawah ini.

func TestWebRolePerm_ListPermissionKeys_Success(t *testing.T) {
	w := testSrv.GET("/api/v1/web/role-permission/permission-keys", superadminHeader())
	assert.Equal(t, 200, w.Code)

	var resp struct {
		Data []struct {
			Key         string `json:"key"`
			DisplayName string `json:"display_name"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	byKey := make(map[string]string, len(resp.Data))
	for _, item := range resp.Data {
		byKey[item.Key] = item.DisplayName
	}
	assert.Equal(t, "Manage System Settings", byKey["manage_system_settings"])
	assert.Equal(t, "Assign Role", byKey["assign_role"])
	assert.Equal(t, "Manage Users", byKey["manage_users"])
}

func TestWebRolePerm_ListPermissionKeys_Unauthenticated(t *testing.T) {
	w := testSrv.GET("/api/v1/web/role-permission/permission-keys")
	assert.Equal(t, 401, w.Code)
}

// createCustomRole adalah helper: bikin role custom baru dan mengembalikan role_id-nya.
func createCustomRole(t *testing.T) string {
	t.Helper()
	w := testSrv.POST("/api/v1/web/role-permission/roles",
		map[string]any{
			"name":         "custom_" + uniqueUsername("role"),
			"display_name": "Custom Role",
			"role_type":    "custom",
			"scope_type":   "global",
			"assignable":   true,
		},
		superadminHeader(),
	)
	require.Equal(t, 201, w.Code, w.Body.String())
	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.NotEmpty(t, resp.Data.ID)
	return resp.Data.ID
}

func TestWebRolePerm_AssignRolePermission_CustomRole_Success(t *testing.T) {
	roleID := createCustomRole(t)

	w := testSrv.POST("/api/v1/web/role-permission/roles/"+roleID+"/permissions",
		map[string]any{"permission_key": "manage_users"},
		superadminHeader(),
	)
	assert.Equal(t, 201, w.Code, w.Body.String())

	var resp struct {
		Data struct {
			Permissions []string `json:"permissions"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Contains(t, resp.Data.Permissions, "manage_users")
}

func TestWebRolePerm_AssignRolePermission_InvalidKey(t *testing.T) {
	roleID := createCustomRole(t)

	w := testSrv.POST("/api/v1/web/role-permission/roles/"+roleID+"/permissions",
		map[string]any{"permission_key": "not_a_real_permission"},
		superadminHeader(),
	)
	assert.True(t, w.Code == 400 || w.Code == 422, "expected 400 or 422, got %d: %s", w.Code, w.Body.String())
}

func TestWebRolePerm_AssignRolePermission_SystemRoleRejected(t *testing.T) {
	var memberRoleID string
	err := testSrv.DB.QueryRowContext(t.Context(), `SELECT id FROM roles WHERE name = 'member' LIMIT 1`).Scan(&memberRoleID)
	require.NoError(t, err)

	w := testSrv.POST("/api/v1/web/role-permission/roles/"+memberRoleID+"/permissions",
		map[string]any{"permission_key": "manage_users"},
		superadminHeader(),
	)
	assert.Equal(t, 409, w.Code, w.Body.String())
}

func TestWebRolePerm_AssignRolePermission_Unauthenticated(t *testing.T) {
	roleID := createCustomRole(t)

	w := testSrv.POST("/api/v1/web/role-permission/roles/"+roleID+"/permissions",
		map[string]any{"permission_key": "manage_users"},
	)
	assert.Equal(t, 401, w.Code)
}

func TestWebRolePerm_RevokeRolePermission_Success(t *testing.T) {
	roleID := createCustomRole(t)

	w := testSrv.POST("/api/v1/web/role-permission/roles/"+roleID+"/permissions",
		map[string]any{"permission_key": "manage_users"},
		superadminHeader(),
	)
	require.Equal(t, 201, w.Code, w.Body.String())

	w = testSrv.DELETE("/api/v1/web/role-permission/roles/"+roleID+"/permissions/manage_users", superadminHeader())
	assert.Equal(t, 200, w.Code, w.Body.String())

	var resp struct {
		Data struct {
			Permissions []string `json:"permissions"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.NotContains(t, resp.Data.Permissions, "manage_users")
}

func TestWebRolePerm_RevokeRolePermission_NotFound(t *testing.T) {
	roleID := createCustomRole(t)

	w := testSrv.DELETE("/api/v1/web/role-permission/roles/"+roleID+"/permissions/manage_users", superadminHeader())
	assert.Equal(t, 404, w.Code, w.Body.String())
}

// TestWebRolePerm_CustomRolePermission_ReflectedInUserSession memverifikasi
// end-to-end: permission yang di-assign ke custom role via API benar-benar
// muncul di sesi user yang diberi role tsb — bukan cuma tersimpan di DB.
func TestWebRolePerm_CustomRolePermission_ReflectedInUserSession(t *testing.T) {
	roleID := createCustomRole(t)

	w := testSrv.POST("/api/v1/web/role-permission/roles/"+roleID+"/permissions",
		map[string]any{"permission_key": "manage_users"},
		superadminHeader(),
	)
	require.Equal(t, 201, w.Code, w.Body.String())

	targetEmail := uniqueEmail("customperm")
	targetID := mustRegisterUser(t, targetEmail, uniqueUsername("cp"), "Secret123!")
	targetToken, _ := mustLogin(t, targetEmail, "Secret123!")

	w = testSrv.POST("/api/v1/web/role-permission/user-roles",
		map[string]any{
			"user_id":    targetID,
			"role_id":    roleID,
			"scope_type": "global",
		},
		superadminHeader(),
	)
	require.Equal(t, 201, w.Code, w.Body.String())

	w = testSrv.GET("/api/v1/auth/session", testhelper.BearerHeader(targetToken))
	assert.Equal(t, 200, w.Code, w.Body.String())

	var session struct {
		Data struct {
			Permissions []struct {
				Key string `json:"key"`
			} `json:"permissions"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&session))
	keys := make([]string, 0, len(session.Data.Permissions))
	for _, p := range session.Data.Permissions {
		keys = append(keys, p.Key)
	}
	assert.Contains(t, keys, "manage_users")
}
