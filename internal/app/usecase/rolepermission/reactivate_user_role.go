package rolepermission

import (
	"context"
	"strings"

	"sipon-api/internal/app/dto"
	rolerepo "sipon-api/internal/domain/role/repository"
)

type ReactivateUserRoleUseCase struct {
	userRoleRepo rolerepo.UserRoleRepository
	getUserRole  *GetUserRoleUseCase
}

func NewReactivateUserRoleUseCase(userRoleRepo rolerepo.UserRoleRepository, getUserRole *GetUserRoleUseCase) *ReactivateUserRoleUseCase {
	return &ReactivateUserRoleUseCase{userRoleRepo: userRoleRepo, getUserRole: getUserRole}
}

// Required — role: superadmin, usergod | perm: assign_role | benefit: -
func (uc *ReactivateUserRoleUseCase) Execute(ctx context.Context, userRoleID string) (*dto.UserRoleResponse, error) {
	userRole, err := uc.userRoleRepo.FindByID(ctx, strings.TrimSpace(userRoleID))
	if err != nil {
		return nil, mapUserRoleDomainError(err)
	}
	userRole.Reactivate()
	if err := uc.userRoleRepo.Update(ctx, userRole); err != nil {
		return nil, mapUserRoleDomainError(err)
	}
	return uc.getUserRole.Execute(ctx, userRole.ID)
}
