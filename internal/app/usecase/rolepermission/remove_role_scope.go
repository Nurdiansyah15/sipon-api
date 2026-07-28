package rolepermission

import (
	"context"
	"strings"

	"sipon-api/internal/app/apperror"
	rolerepo "sipon-api/internal/domain/role/repository"
)

type RemoveRoleScopeUseCase struct {
	roleScopeRepo rolerepo.RoleScopeRepository
}

func NewRemoveRoleScopeUseCase(roleScopeRepo rolerepo.RoleScopeRepository) *RemoveRoleScopeUseCase {
	return &RemoveRoleScopeUseCase{roleScopeRepo: roleScopeRepo}
}

func (uc *RemoveRoleScopeUseCase) Execute(ctx context.Context, scopeID string) error {
	scopeID = strings.TrimSpace(scopeID)
	if scopeID == "" {
		return apperror.Unprocessable("scope_id wajib diisi", nil)
	}
	_, err := uc.roleScopeRepo.FindByID(ctx, scopeID)
	if err != nil {
		return apperror.NotFound("role scope tidak ditemukan", err)
	}
	return uc.roleScopeRepo.Delete(ctx, scopeID)
}
