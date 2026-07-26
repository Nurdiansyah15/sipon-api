package rolepermission

import (
	"context"
	"strings"

	"sipon-api/internal/app/dto"
	rolerepo "sipon-api/internal/domain/role/repository"
)

type UpdateRoleUseCase struct {
	roleRepo           rolerepo.RoleRepository
	rolePermissionRepo rolerepo.RolePermissionRepository
}

func NewUpdateRoleUseCase(roleRepo rolerepo.RoleRepository, rolePermissionRepo rolerepo.RolePermissionRepository) *UpdateRoleUseCase {
	return &UpdateRoleUseCase{roleRepo: roleRepo, rolePermissionRepo: rolePermissionRepo}
}

// Required — role: superadmin, usergod | perm: - | benefit: -
func (uc *UpdateRoleUseCase) Execute(ctx context.Context, roleID string, req dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	role, err := uc.roleRepo.FindByID(ctx, strings.TrimSpace(roleID))
	if err != nil {
		return nil, mapRoleDomainError(err)
	}

	if err := role.UpdateDetails(req.DisplayName, req.Description, req.Assignable); err != nil {
		return nil, mapRoleDomainValidationError(err)
	}

	if err := uc.roleRepo.Update(ctx, role); err != nil {
		return nil, mapRoleDomainError(err)
	}
	return buildRoleResponse(ctx, uc.roleRepo, uc.rolePermissionRepo, role.ID)
}
