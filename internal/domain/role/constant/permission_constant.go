package constant

import "sort"

// PermissionKey didefinisikan sebagai constant di kode, bukan tabel master DB —
// menambah permission baru berarti menambah constant di sini (bersamaan dengan
// guard middleware.RequirePermission yang memakainya), bukan lewat API/migrasi.
//
// Mapping permission→role SYSTEM (RolePermissions di bawah) juga tetap fixed di
// kode. Untuk role CUSTOM, permission-nya BOLEH diatur dinamis oleh user lewat
// API (disimpan di tabel role_permissions) — tapi permission_key yang boleh
// di-assign tetap harus salah satu dari AllPermissionKeys() di sini, supaya
// tidak ada key sembarangan yang tidak punya guard apa pun di kode.
type PermissionKey string

const (
	PermissionManageSystemSettings   PermissionKey = "manage_system_settings"
	PermissionAssignRole             PermissionKey = "assign_role"
	PermissionManageUsers            PermissionKey = "manage_users"
	PermissionResetUserPassword      PermissionKey = "reset_user_password"
	PermissionDeactivateUser          PermissionKey = "deactivate_user"
	PermissionManageRoles            PermissionKey = "manage_roles"
	PermissionManageRolePermissions   PermissionKey = "manage_role_permissions"
)

// PermissionDefinition adalah metadata satu permission (dipakai sebagai bentuk
// return AllPermissionDefinitions — lihat DefaultPermissionsInit untuk sumber datanya).
type PermissionDefinition struct {
	Key         PermissionKey
	DisplayName string
	Description string
}

// PermissionInit adalah tipe pendaftaran metadata permission — analog dengan
// RoleInit untuk role. Key harus salah satu PermissionKey constant di atas.
type PermissionInit map[PermissionKey]struct {
	DisplayName string
	Description string
}

// DefaultPermissionsInit adalah SATU-SATUNYA tempat mendaftarkan permission
// beserta metadata-nya (display name & description untuk admin UI). Menambah
// permission baru = menambah constant PermissionKey di atas + entry di sini +
// (kalau relevan) guard middleware.RequirePermission di route registration +
// entry di RolePermissions untuk system role yang berhak.
var DefaultPermissionsInit = PermissionInit{
	PermissionManageSystemSettings: {
		DisplayName: "Manage System Settings",
		Description: "Mengelola pengaturan sistem global.",
	},
	PermissionAssignRole: {
		DisplayName: "Assign Role",
		Description: "Menetapkan atau mencabut role pada user.",
	},
	PermissionManageUsers: {
		DisplayName: "Manage Users",
		Description: "Mengelola akun user (lihat, ubah status, dsb).",
	},
	PermissionResetUserPassword: {
		DisplayName: "Reset User Password",
		Description: "Menyetel ulang kata sandi user lain (admin-generate temp password).",
	},
	PermissionDeactivateUser: {
		DisplayName: "Deactivate User",
		Description: "Menonaktifkan atau mengaktifkan kembali akun user.",
	},
	PermissionManageRoles: {
		DisplayName: "Manage Roles",
		Description: "Membuat dan mengubah metadata role (definisi role).",
	},
	PermissionManageRolePermissions: {
		DisplayName: "Manage Role Permissions",
		Description: "Menetapkan atau mencabut permission pada role custom.",
	},
}

// AllPermissionDefinitions mengembalikan seluruh permission dari
// DefaultPermissionsInit beserta metadata-nya, terurut berdasarkan key supaya
// urutannya stabil antar pemanggilan (map di Go tidak menjamin urutan) —
// dipakai endpoint katalog GET /role-permission/permission-keys.
func AllPermissionDefinitions() []PermissionDefinition {
	defs := make([]PermissionDefinition, 0, len(DefaultPermissionsInit))
	for key, meta := range DefaultPermissionsInit {
		defs = append(defs, PermissionDefinition{
			Key:         key,
			DisplayName: meta.DisplayName,
			Description: meta.Description,
		})
	}
	sort.Slice(defs, func(i, j int) bool { return defs[i].Key < defs[j].Key })
	return defs
}

// AllPermissionKeys mengembalikan seluruh permission key yang terdaftar di
// DefaultPermissionsInit. Dipakai untuk validasi saat assign permission ke
// custom role (lihat IsValidPermissionKey).
func AllPermissionKeys() []PermissionKey {
	keys := make([]PermissionKey, 0, len(DefaultPermissionsInit))
	for key := range DefaultPermissionsInit {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

// IsValidPermissionKey mengecek apakah key terdaftar di DefaultPermissionsInit.
func IsValidPermissionKey(key PermissionKey) bool {
	_, ok := DefaultPermissionsInit[key]
	return ok
}

// RolePermissions memetakan setiap system role ke daftar permission yang
// dimilikinya (fixed, tidak bisa diubah lewat API — lihat AssignRolePermissionUseCase
// yang menolak assignment untuk role bertipe system).
var RolePermissions = map[RoleName][]PermissionKey{
	UserGodRoleName: {
		PermissionManageSystemSettings, PermissionAssignRole, PermissionManageUsers,
		PermissionResetUserPassword, PermissionDeactivateUser,
		PermissionManageRoles, PermissionManageRolePermissions,
	},
	SuperAdminRoleName: {
		PermissionManageSystemSettings, PermissionAssignRole, PermissionManageUsers,
		PermissionResetUserPassword, PermissionDeactivateUser,
		PermissionManageRoles, PermissionManageRolePermissions,
	},
	AdminRoleName: {
		PermissionAssignRole, PermissionManageUsers,
		PermissionResetUserPassword, PermissionDeactivateUser,
	},
	MemberRoleName: {},
}

// PermissionsForRole mengembalikan daftar permission key system-role dari constant ini.
// Untuk role custom, permission-nya berasal dari tabel role_permissions — lihat
// RolePermissionRepository, bukan fungsi ini.
func PermissionsForRole(name RoleName) []PermissionKey {
	return RolePermissions[name]
}

// RoleHasPermission mengecek apakah sebuah system role memiliki permission tertentu.
func RoleHasPermission(name RoleName, key PermissionKey) bool {
	for _, k := range RolePermissions[name] {
		if k == key {
			return true
		}
	}
	return false
}
