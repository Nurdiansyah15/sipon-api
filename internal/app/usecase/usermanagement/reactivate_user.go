package usermanagement

import (
	"context"
	"strings"

	"sipon-api/internal/app/dto"
	userrepo "sipon-api/internal/domain/user/repository"
)

type ReactivateUserUseCase struct {
	userRepo userrepo.UserRepository
}

func NewReactivateUserUseCase(userRepo userrepo.UserRepository) *ReactivateUserUseCase {
	return &ReactivateUserUseCase{userRepo: userRepo}
}

// Required — perm: deactivate_user
//
// Mengembalikan akun user ke StatusActive. Menolak double-reactivate
// (CodeUserAlreadyActive → 409). Sama permission dengan Deactivate (lihat
// plan §constant: deactivate_user covers both directions).
func (uc *ReactivateUserUseCase) Execute(ctx context.Context, userID string) (*dto.UserManagementResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		return nil, mapUserDomainError(err)
	}
	if err := user.Reactivate(); err != nil {
		return nil, mapUserDomainError(err)
	}
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, mapUserDomainError(err)
	}
	return buildUserManagementResponse(ctx, nil, user, false)
}