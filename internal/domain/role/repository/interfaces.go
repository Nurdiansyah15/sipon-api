package repository

import (
	"context"
	"sipon-api/internal/domain/role/constant"
	"sipon-api/internal/domain/role/entity"
)

// RoleRepository menyimpan master role beserta perubahan metadata.
type RoleRepository interface {
	Save(ctx context.Context, role *entity.Role) error
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*entity.Role, error)
	FindByName(ctx context.Context, name constant.RoleName) (*entity.Role, error)
	ListByType(ctx context.Context, roleType constant.RoleType) ([]*entity.Role, error)
}

// UserRoleRepository mengelola assignment role user lintas scope.
type UserRoleRepository interface {
	Save(ctx context.Context, userRole *entity.UserRole) error
	Update(ctx context.Context, userRole *entity.UserRole) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*entity.UserRole, error)
	FindActiveByUserID(ctx context.Context, userID string) ([]*entity.UserRole, error)
	// ListActiveUserIDsByRoleName mengembalikan user_id yang punya assignment
	// role aktif (belum expired) dengan nama role tsb.
	ListActiveUserIDsByRoleName(ctx context.Context, roleName constant.RoleName) ([]string, error)
}

// RolePermissionRepository mengelola assignment permission untuk role CUSTOM.
// Role system tidak pernah disimpan di sini — permission-nya fixed dari
// constant.RolePermissions (lihat Role.HasPermission).
type RolePermissionRepository interface {
	Save(ctx context.Context, rp *entity.RolePermission) error
	Delete(ctx context.Context, roleID string, permissionKey constant.PermissionKey) error
	ListByRoleID(ctx context.Context, roleID string) ([]*entity.RolePermission, error)
}
