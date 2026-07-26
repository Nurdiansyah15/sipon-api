package rolepermission

import (
	"context"
	"strings"

	rolerepo "sipon-api/internal/domain/role/repository"
)

type DeleteUserRoleUseCase struct{ userRoleRepo rolerepo.UserRoleRepository }

func NewDeleteUserRoleUseCase(userRoleRepo rolerepo.UserRoleRepository) *DeleteUserRoleUseCase {
	return &DeleteUserRoleUseCase{userRoleRepo: userRoleRepo}
}

// Required — role: superadmin, usergod | perm: assign_role | benefit: -
func (uc *DeleteUserRoleUseCase) Execute(ctx context.Context, userRoleID string) error {
	if err := uc.userRoleRepo.Delete(ctx, strings.TrimSpace(userRoleID)); err != nil {
		return mapUserRoleDomainError(err)
	}
	return nil
}
