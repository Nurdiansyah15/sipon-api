package rolepermission

import (
	"context"
	"strings"

	"sipon-api/internal/app/dto"
	rolerepo "sipon-api/internal/domain/role/repository"
)

type AssignUserRoleUseCase struct {
	roleRepo     rolerepo.RoleRepository
	userRoleRepo rolerepo.UserRoleRepository
	getUserRole  *GetUserRoleUseCase
}

func NewAssignUserRoleUseCase(roleRepo rolerepo.RoleRepository, userRoleRepo rolerepo.UserRoleRepository, getUserRole *GetUserRoleUseCase) *AssignUserRoleUseCase {
	return &AssignUserRoleUseCase{roleRepo: roleRepo, userRoleRepo: userRoleRepo, getUserRole: getUserRole}
}

// Required — role: superadmin, usergod | perm: assign_role | benefit: -
func (uc *AssignUserRoleUseCase) Execute(ctx context.Context, actorUserID string, req dto.AssignUserRoleRequest) (*dto.UserRoleResponse, error) {
	role, err := uc.roleRepo.FindByID(ctx, strings.TrimSpace(req.RoleID))
	if err != nil {
		return nil, mapRoleDomainError(err)
	}
	if err := role.EnsureAssignable(); err != nil {
		return nil, mapRoleDomainError(err)
	}
	scopeType, err := parseScopeType(req.ScopeType, true)
	if err != nil {
		return nil, err
	}
	if err := role.EnsureAssignmentScopeMatch(scopeType); err != nil {
		return nil, mapRoleDomainError(err)
	}
	assignment, err := newUserRoleAssignment(actorUserID, req, role.ID)
	if err != nil {
		return nil, err
	}
	if err := uc.userRoleRepo.Save(ctx, assignment); err != nil {
		return nil, mapUserRoleDomainError(err)
	}
	return uc.getUserRole.Execute(ctx, assignment.ID)
}
