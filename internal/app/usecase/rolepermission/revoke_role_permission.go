package rolepermission

import (
	"context"
	"strings"

	"sipon-api/internal/app/dto"
	roleconstant "sipon-api/internal/domain/role/constant"
	rolerepo "sipon-api/internal/domain/role/repository"
)

type RevokeRolePermissionUseCase struct {
	roleRepo           rolerepo.RoleRepository
	rolePermissionRepo rolerepo.RolePermissionRepository
}

func NewRevokeRolePermissionUseCase(roleRepo rolerepo.RoleRepository, rolePermissionRepo rolerepo.RolePermissionRepository) *RevokeRolePermissionUseCase {
	return &RevokeRolePermissionUseCase{roleRepo: roleRepo, rolePermissionRepo: rolePermissionRepo}
}

// Required — role: superadmin, usergod | perm: manage_system_settings | benefit: -
func (uc *RevokeRolePermissionUseCase) Execute(ctx context.Context, roleID, permissionKey string) (*dto.RoleResponse, error) {
	role, err := uc.roleRepo.FindByID(ctx, strings.TrimSpace(roleID))
	if err != nil {
		return nil, mapRoleDomainError(err)
	}
	if err := role.EnsureCustom(); err != nil {
		return nil, mapRolePermissionDomainError(err)
	}

	if err := uc.rolePermissionRepo.Delete(ctx, role.ID, roleconstant.PermissionKey(strings.TrimSpace(permissionKey))); err != nil {
		return nil, mapRolePermissionDomainError(err)
	}

	return buildRoleResponse(ctx, uc.roleRepo, uc.rolePermissionRepo, role.ID)
}
