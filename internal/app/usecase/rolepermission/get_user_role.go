package rolepermission

import (
	"context"

	"sipon-api/internal/app/dto"
	rolerepo "sipon-api/internal/domain/role/repository"
	userrepo "sipon-api/internal/domain/user/repository"
)

type GetUserRoleUseCase struct {
	userRepo           userrepo.UserRepository
	roleRepo           rolerepo.RoleRepository
	userRoleRepo       rolerepo.UserRoleRepository
	rolePermissionRepo rolerepo.RolePermissionRepository
}

func NewGetUserRoleUseCase(userRepo userrepo.UserRepository, roleRepo rolerepo.RoleRepository, userRoleRepo rolerepo.UserRoleRepository, rolePermissionRepo rolerepo.RolePermissionRepository) *GetUserRoleUseCase {
	return &GetUserRoleUseCase{userRepo: userRepo, roleRepo: roleRepo, userRoleRepo: userRoleRepo, rolePermissionRepo: rolePermissionRepo}
}

// Required — role: superadmin, usergod | perm: - | benefit: -
func (uc *GetUserRoleUseCase) Execute(ctx context.Context, userRoleID string) (*dto.UserRoleResponse, error) {
	return buildUserRoleResponse(ctx, uc.userRepo, uc.roleRepo, uc.userRoleRepo, uc.rolePermissionRepo, userRoleID)
}
