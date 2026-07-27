package dto

import "time"

// ListUsersQuery adalah filter untuk admin user listing endpoint.
type ListUsersQuery struct {
	Status string `form:"status"`
	RoleID string `form:"role_id"`
	Search string `form:"search"`
	PaginationParams
}

// UserRoleSummaryResponse ringkasan assignment role aktif untuk satu user.
type UserRoleSummaryResponse struct {
	ID        string  `json:"id"`
	RoleID    string  `json:"role_id"`
	RoleName  string  `json:"role_name"`
	ScopeType string  `json:"scope_type"`
	ScopeID   *string `json:"scope_id,omitempty"`
	IsActive  bool    `json:"is_active"`
}

// UserManagementResponse — pada listing, Roles dikosongkan untuk menghindari
// N+1 (lihat docs/plans/system-management-module.md §6). Pada get-by-id, diisi.
type UserManagementResponse struct {
	ID          string                   `json:"id"`
	Username    string                   `json:"username"`
	Fullname    *string                  `json:"fullname,omitempty"`
	Email       string                   `json:"email"`
	Phone       *string                  `json:"phone,omitempty"`
	Status      string                   `json:"status"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	LastLoginAt *time.Time               `json:"last_login_at,omitempty"`
	Roles       []UserRoleSummaryResponse `json:"roles,omitempty"`
}

// CreateUserRequest payload admin untuk membuat user baru. Password
// auto-generated (lihat docs/plans/system-management-module.md §Decision 1).
type CreateUserRequest struct {
	Username string  `json:"username" binding:"required"`
	Fullname *string `json:"fullname,omitempty"`
	Email    string  `json:"email" binding:"required"`
	Phone    *string `json:"phone,omitempty"`
}

// CreateUserResponse — generated_password ditampilkan sekali saja. Frontend
// harus menyajikan dengan tombol copy + peringatan "won't be shown again".
type CreateUserResponse struct {
	UserManagementResponse
	GeneratedPassword string `json:"generated_password"`
}

// ResetUserPasswordResponse — generated_password baru, ditampilkan sekali.
type ResetUserPasswordResponse struct {
	GeneratedPassword string `json:"generated_password"`
}