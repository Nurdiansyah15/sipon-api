package rolepermission

import (
	"context"

	"sipon-api/internal/app/dto"
	rolerepo "sipon-api/internal/domain/role/repository"
)

type GetRoleUseCase struct {
	roleRepo           rolerepo.RoleRepository
	rolePermissionRepo rolerepo.RolePermissionRepository
}

func NewGetRoleUseCase(roleRepo rolerepo.RoleRepository, rolePermissionRepo rolerepo.RolePermissionRepository) *GetRoleUseCase {
	return &GetRoleUseCase{roleRepo: roleRepo, rolePermissionRepo: rolePermissionRepo}
}

// Required — role: superadmin, usergod | perm: - | benefit: -
func (uc *GetRoleUseCase) Execute(ctx context.Context, roleID string) (*dto.RoleResponse, error) {
	return buildRoleResponse(ctx, uc.roleRepo, uc.rolePermissionRepo, roleID)
}
