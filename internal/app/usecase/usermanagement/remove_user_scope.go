package usermanagement

import (
	"context"
	"strings"

	"sipon-api/internal/app/apperror"
	userrepo "sipon-api/internal/domain/user/repository"
)

type RemoveUserScopeUseCase struct {
	userScopeRepo userrepo.UserScopeRepository
}

func NewRemoveUserScopeUseCase(userScopeRepo userrepo.UserScopeRepository) *RemoveUserScopeUseCase {
	return &RemoveUserScopeUseCase{userScopeRepo: userScopeRepo}
}

func (uc *RemoveUserScopeUseCase) Execute(ctx context.Context, scopeID string) error {
	scopeID = strings.TrimSpace(scopeID)
	if scopeID == "" {
		return apperror.Unprocessable("scope_id wajib diisi", nil)
	}

	_, err := uc.userScopeRepo.FindByID(ctx, scopeID)
	if err != nil {
		return apperror.NotFound("user scope tidak ditemukan", err)
	}

	return uc.userScopeRepo.Delete(ctx, scopeID)
}
