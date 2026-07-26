package dto

import "time"

type ListRolesQuery struct {
	RoleType   string `form:"role_type"`
	ScopeType  string `form:"scope_type"`
	Assignable string `form:"assignable"`
	PaginationParams
}

type CreateRoleRequest struct {
	Name        string  `json:"name" binding:"required"`
	DisplayName string  `json:"display_name" binding:"required"`
	Description *string `json:"description,omitempty"`
	RoleType    string  `json:"role_type" binding:"required,oneof=system custom"`
	ScopeType   string  `json:"scope_type" binding:"required,oneof=global region community"`
	Assignable  bool    `json:"assignable"`
}

type UpdateRoleRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Description *string `json:"description,omitempty"`
	Assignable  *bool   `json:"assignable,omitempty"`
}

type RoleSummaryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	RoleType    string `json:"role_type,omitempty"`
	Assignable  bool   `json:"assignable,omitempty"`
}

// RoleResponse.Permissions berisi permission key (string). Untuk role system,
// dihitung dari constant.PermissionsForRole(role.Name); untuk role custom,
// dari tabel role_permissions (lihat internal/domain/role/constant dan
// RolePermissionRepository).
type RoleResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description *string   `json:"description,omitempty"`
	RoleType    string    `json:"role_type"`
	ScopeType   string    `json:"scope_type"`
	Assignable  bool      `json:"assignable"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Permissions []string  `json:"permissions,omitempty"`
}

type ListUserRolesQuery struct {
	UserID    string `form:"user_id"`
	RoleID    string `form:"role_id"`
	ScopeType string `form:"scope_type"`
	ScopeID   string `form:"scope_id"`
	IsActive  string `form:"is_active"`
	PaginationParams
}

type AssignUserRoleRequest struct {
	UserID    string     `json:"user_id" binding:"required"`
	RoleID    string     `json:"role_id" binding:"required"`
	ScopeType string     `json:"scope_type" binding:"required,oneof=global region community"`
	ScopeID   *string    `json:"scope_id,omitempty"`
	ExpiredAt *time.Time `json:"expired_at,omitempty"`
	Notes     *string    `json:"notes,omitempty"`
}

type UpdateUserRoleRequest struct {
	ExpiredAt *time.Time `json:"expired_at,omitempty"`
}

// AssignRolePermissionRequest — assign satu permission ke role CUSTOM.
// permission_key harus salah satu dari GET /role-permission/permission-keys.
type AssignRolePermissionRequest struct {
	PermissionKey string  `json:"permission_key" binding:"required"`
	Notes         *string `json:"notes,omitempty"`
}

// PermissionKeyResponse adalah satu entri katalog permission yang dikenal sistem
// (lihat constant.AllPermissionDefinitions) — dipakai frontend untuk menampilkan
// pilihan permission saat assign ke custom role, tanpa hardcode di client.
type PermissionKeyResponse struct {
	Key         string `json:"key"`
	DisplayName string `json:"display_name"`
	Description string `json:"description,omitempty"`
}

type UserSummaryResponse struct {
	ID    string  `json:"id"`
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
	Phone *string `json:"phone,omitempty"`
}

type UserRoleResponse struct {
	ID            string              `json:"id"`
	UserID        string              `json:"user_id"`
	User          UserSummaryResponse `json:"user"`
	RoleID        string              `json:"role_id"`
	Role          RoleSummaryResponse `json:"role"`
	ScopeType     string              `json:"scope_type"`
	ScopeID       *string             `json:"scope_id,omitempty"`
	AssignedAt    time.Time           `json:"assigned_at"`
	AssignedBy    *string             `json:"assigned_by,omitempty"`
	ExpiredAt     *time.Time          `json:"expired_at,omitempty"`
	IsActive      bool                `json:"is_active"`
	DeactivatedAt *time.Time          `json:"deactivated_at,omitempty"`
	Permissions   []string            `json:"permissions,omitempty"`
}
