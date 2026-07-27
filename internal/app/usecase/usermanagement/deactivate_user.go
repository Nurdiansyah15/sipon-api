package usermanagement

import (
	"context"
	"strings"

	"sipon-api/internal/app/dto"
	userrepo "sipon-api/internal/domain/user/repository"
)

type DeactivateUserUseCase struct {
	userRepo userrepo.UserRepository
}

func NewDeactivateUserUseCase(userRepo userrepo.UserRepository) *DeactivateUserUseCase {
	return &DeactivateUserUseCase{userRepo: userRepo}
}

// Required — perm: deactivate_user
//
// Mem-banned akun user (reuses StatusBanned). Menolak double-deactivate
// (CodeUserAlreadyBanned → 409) supaya idempotency terlihat jelas.
func (uc *DeactivateUserUseCase) Execute(ctx context.Context, userID string) (*dto.UserManagementResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		return nil, mapUserDomainError(err)
	}
	if err := user.Deactivate(); err != nil {
		return nil, mapUserDomainError(err)
	}
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, mapUserDomainError(err)
	}
	return buildUserManagementResponse(ctx, nil, user, false)
}