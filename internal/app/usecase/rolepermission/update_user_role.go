package rolepermission

import (
	"context"
	"strings"

	"sipon-api/internal/app/dto"
	rolerepo "sipon-api/internal/domain/role/repository"
)

type UpdateUserRoleUseCase struct {
	userRoleRepo rolerepo.UserRoleRepository
	getUserRole  *GetUserRoleUseCase
}

func NewUpdateUserRoleUseCase(userRoleRepo rolerepo.UserRoleRepository, getUserRole *GetUserRoleUseCase) *UpdateUserRoleUseCase {
	return &UpdateUserRoleUseCase{userRoleRepo: userRoleRepo, getUserRole: getUserRole}
}

// Required — role: superadmin, usergod | perm: assign_role | benefit: -
func (uc *UpdateUserRoleUseCase) Execute(ctx context.Context, userRoleID string, req dto.UpdateUserRoleRequest) (*dto.UserRoleResponse, error) {
	userRole, err := uc.userRoleRepo.FindByID(ctx, strings.TrimSpace(userRoleID))
	if err != nil {
		return nil, mapUserRoleDomainError(err)
	}
	userRole.UpdateExpiration(req.ExpiredAt)
	if err := uc.userRoleRepo.Update(ctx, userRole); err != nil {
		return nil, mapUserRoleDomainError(err)
	}
	return uc.getUserRole.Execute(ctx, userRole.ID)
}
