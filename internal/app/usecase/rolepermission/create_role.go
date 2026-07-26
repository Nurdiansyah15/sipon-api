package rolepermission

import (
	"context"

	"sipon-api/internal/app/dto"
	rolerepo "sipon-api/internal/domain/role/repository"
)

type CreateRoleUseCase struct {
	roleRepo           rolerepo.RoleRepository
	rolePermissionRepo rolerepo.RolePermissionRepository
}

func NewCreateRoleUseCase(roleRepo rolerepo.RoleRepository, rolePermissionRepo rolerepo.RolePermissionRepository) *CreateRoleUseCase {
	return &CreateRoleUseCase{roleRepo: roleRepo, rolePermissionRepo: rolePermissionRepo}
}

// Required — role: superadmin, usergod | perm: manage_system_settings | benefit: -
func (uc *CreateRoleUseCase) Execute(ctx context.Context, req dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	role, err := newRoleEntity(req)
	if err != nil {
		return nil, err
	}
	if err := uc.roleRepo.Save(ctx, role); err != nil {
		return nil, mapRoleDomainError(err)
	}
	return buildRoleResponse(ctx, uc.roleRepo, uc.rolePermissionRepo, role.ID)
}
