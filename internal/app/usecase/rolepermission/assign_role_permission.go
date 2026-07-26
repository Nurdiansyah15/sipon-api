package rolepermission

import (
	"context"
	"strings"

	"sipon-api/internal/app/dto"
	roleconstant "sipon-api/internal/domain/role/constant"
	roleentity "sipon-api/internal/domain/role/entity"
	rolerepo "sipon-api/internal/domain/role/repository"

	"github.com/google/uuid"
)

type AssignRolePermissionUseCase struct {
	roleRepo           rolerepo.RoleRepository
	rolePermissionRepo rolerepo.RolePermissionRepository
}

func NewAssignRolePermissionUseCase(roleRepo rolerepo.RoleRepository, rolePermissionRepo rolerepo.RolePermissionRepository) *AssignRolePermissionUseCase {
	return &AssignRolePermissionUseCase{roleRepo: roleRepo, rolePermissionRepo: rolePermissionRepo}
}

// Required — role: superadmin, usergod | perm: manage_system_settings | benefit: -
// Hanya berlaku untuk role custom — permission role system fixed di constant.
func (uc *AssignRolePermissionUseCase) Execute(ctx context.Context, actorUserID, roleID string, req dto.AssignRolePermissionRequest) (*dto.RoleResponse, error) {
	role, err := uc.roleRepo.FindByID(ctx, strings.TrimSpace(roleID))
	if err != nil {
		return nil, mapRoleDomainError(err)
	}
	if err := role.EnsureCustom(); err != nil {
		return nil, mapRolePermissionDomainError(err)
	}

	assignment, err := roleentity.NewRolePermission(
		uuid.NewString(), role.ID,
		roleconstant.PermissionKey(strings.TrimSpace(req.PermissionKey)),
		strings.TrimSpace(actorUserID), req.Notes,
	)
	if err != nil {
		return nil, mapRolePermissionDomainError(err)
	}

	if err := uc.rolePermissionRepo.Save(ctx, assignment); err != nil {
		return nil, mapRolePermissionDomainError(err)
	}

	return buildRoleResponse(ctx, uc.roleRepo, uc.rolePermissionRepo, role.ID)
}
