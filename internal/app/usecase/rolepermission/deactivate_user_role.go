package rolepermission

import (
	"context"
	"strings"
	"time"

	"sipon-api/internal/app/dto"
	rolerepo "sipon-api/internal/domain/role/repository"
)

type DeactivateUserRoleUseCase struct {
	userRoleRepo rolerepo.UserRoleRepository
	getUserRole  *GetUserRoleUseCase
}

func NewDeactivateUserRoleUseCase(userRoleRepo rolerepo.UserRoleRepository, getUserRole *GetUserRoleUseCase) *DeactivateUserRoleUseCase {
	return &DeactivateUserRoleUseCase{userRoleRepo: userRoleRepo, getUserRole: getUserRole}
}

// Required — role: superadmin, usergod | perm: assign_role | benefit: -
func (uc *DeactivateUserRoleUseCase) Execute(ctx context.Context, userRoleID string) (*dto.UserRoleResponse, error) {
	userRole, err := uc.userRoleRepo.FindByID(ctx, strings.TrimSpace(userRoleID))
	if err != nil {
		return nil, mapUserRoleDomainError(err)
	}
	userRole.Deactivate(time.Now())
	if err := uc.userRoleRepo.Update(ctx, userRole); err != nil {
		return nil, mapUserRoleDomainError(err)
	}
	return uc.getUserRole.Execute(ctx, userRole.ID)
}
