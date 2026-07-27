package port

import (
	"context"
	"time"

	"sipon-api/internal/app/dto"
)

// UserListReadQuery adalah filter untuk listing user admin (pagination +
// filter status/role/search). role_id memfilter user yang punya assignment
// role aktif tertentu; search melakukan ILIKE pada username/fullname/email.
type UserListReadQuery struct {
	Status  string
	RoleID  string
	Search  string
	dto.PaginationParams
}

// UserReadItem adalah representasi flat satu user untuk admin listing/detail.
// Roles diisi sebagai ringkasan role assignment aktif (di-get-by-id) atau
// dikosongkan pada list untuk menghindari N+1 (lihat docs/plans/system-management-module.md §6).
type UserReadItem struct {
	ID          string
	Username    string
	Fullname    *string
	Email       string
	Phone       *string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastLoginAt *time.Time
}

// UserRoleSummaryReadItem adalah ringkasan assignment role aktif untuk user.
type UserRoleSummaryReadItem struct {
	ID        string
	RoleID    string
	RoleName  string
	ScopeType string
	ScopeID   *string
	IsActive  bool
}

// UserQueryReadModel adalah query model untuk listing/detail user admin.
// Berada di layer port (bukan domain repository) karena listing adalah concern
// read-model per CLAUDE.md §6 — bayangan dari rolepermission.RolePermissionQueryReadModel.
type UserQueryReadModel interface {
	ListUsers(ctx context.Context, query UserListReadQuery) ([]UserReadItem, dto.Meta, error)
	// ListActiveRoleSummariesByUserID mengembalikan assignment role aktif user.
	ListActiveRoleSummariesByUserID(ctx context.Context, userID string) ([]UserRoleSummaryReadItem, error)
	// ListActiveRoleSummariesByUserIDs batch query untuk banyak user sekaligus,
	// menghindari N+1 pada user listing. Mengembalikan map[userID][]summary.
	ListActiveRoleSummariesByUserIDs(ctx context.Context, userIDs []string) (map[string][]UserRoleSummaryReadItem, error)
}