package usermanagement

import (
	"context"
	"strings"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	userrepo "sipon-api/internal/domain/user/repository"
)

type ListUserScopesUseCase struct {
	userScopeRepo userrepo.UserScopeRepository
}

func NewListUserScopesUseCase(userScopeRepo userrepo.UserScopeRepository) *ListUserScopesUseCase {
	return &ListUserScopesUseCase{userScopeRepo: userScopeRepo}
}

func (uc *ListUserScopesUseCase) Execute(ctx context.Context, userID string) ([]dto.UserScopeResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, apperror.Unprocessable("user_id wajib diisi", nil)
	}

	scopes, err := uc.userScopeRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal("gagal mengambil user scopes", err)
	}

	result := make([]dto.UserScopeResponse, 0, len(scopes))
	for _, s := range scopes {
		result = append(result, dto.UserScopeResponse{
			ID:         s.ID,
			ScopeType:  string(s.ScopeType),
			ScopeValue: s.ScopeValue,
		})
	}
	return result, nil
}
