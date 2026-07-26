package port

import (
	"context"
	"sipon-api/internal/app/dto"
	"time"
)

type RoleListReadQuery struct {
	RoleType   string
	ScopeType  string
	Assignable *bool
	dto.PaginationParams
}

type RoleSummaryReadItem struct {
	ID          string
	Name        string
	DisplayName string
	RoleType    string
	Assignable  bool
}

type RoleReadItem struct {
	ID          string
	Name        string
	DisplayName string
	Description *string
	RoleType    string
	ScopeType   string
	Assignable  bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UserRoleListReadQuery struct {
	UserID    string
	RoleID    string
	ScopeType string
	ScopeID   string
	IsActive  *bool
	dto.PaginationParams
}

type UserSummaryReadItem struct {
	ID    string
	Name  *string
	Email *string
	Phone *string
}

type UserRoleReadItem struct {
	ID            string
	UserID        string
	User          UserSummaryReadItem
	RoleID        string
	Role          RoleSummaryReadItem
	ScopeType     string
	ScopeID       *string
	AssignedAt    time.Time
	AssignedBy    *string
	ExpiredAt     *time.Time
	IsActive      bool
	DeactivatedAt *time.Time
	// Permissions berasal dari constant.PermissionsForRole(role.Name) — bukan
	// dari tabel, karena permission tidak disimpan di DB.
	Permissions []string
}

// RolePermissionQueryReadModel adalah query model untuk listing/pagination role
// dan user-role assignment. Tidak ada query permission/role_permission karena
// permission bukan aggregate DB — lihat internal/domain/role/constant.
type RolePermissionQueryReadModel interface {
	ListRoles(ctx context.Context, query RoleListReadQuery) ([]RoleReadItem, dto.Meta, error)
	ListUserRoles(ctx context.Context, query UserRoleListReadQuery) ([]UserRoleReadItem, dto.Meta, error)
}
