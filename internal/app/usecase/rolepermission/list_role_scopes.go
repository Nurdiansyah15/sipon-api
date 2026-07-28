package rolepermission

import (
	"context"
	"strings"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	rolerepo "sipon-api/internal/domain/role/repository"
)

type ListRoleScopesUseCase struct {
	roleScopeRepo rolerepo.RoleScopeRepository
}

func NewListRoleScopesUseCase(roleScopeRepo rolerepo.RoleScopeRepository) *ListRoleScopesUseCase {
	return &ListRoleScopesUseCase{roleScopeRepo: roleScopeRepo}
}

func (uc *ListRoleScopesUseCase) Execute(ctx context.Context, roleID string) ([]dto.RoleScopeResponse, error) {
	roleID = strings.TrimSpace(roleID)
	if roleID == "" {
		return nil, apperror.Unprocessable("role_id wajib diisi", nil)
	}
	scopes, err := uc.roleScopeRepo.FindByRoleID(ctx, roleID)
	if err != nil {
		return nil, apperror.Internal("gagal mengambil role scopes", err)
	}
	result := make([]dto.RoleScopeResponse, 0, len(scopes))
	for _, s := range scopes {
		result = append(result, dto.RoleScopeResponse{
			ID: s.ID, ScopeType: string(s.ScopeType), ScopeValue: s.ScopeValue,
		})
	}
	return result, nil
}
