package usermanagement

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/domain/user/entity"
	userrepo "sipon-api/internal/domain/user/repository"
	"sipon-api/internal/domain/user/valueobject"
)

type AssignUserScopeUseCase struct {
	userRepo      userrepo.UserRepository
	userScopeRepo userrepo.UserScopeRepository
}

func NewAssignUserScopeUseCase(
	userRepo userrepo.UserRepository,
	userScopeRepo userrepo.UserScopeRepository,
) *AssignUserScopeUseCase {
	return &AssignUserScopeUseCase{userRepo: userRepo, userScopeRepo: userScopeRepo}
}

func (uc *AssignUserScopeUseCase) Execute(ctx context.Context, userID string, req dto.AssignUserScopeRequest) (*dto.UserScopeResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, apperror.Unprocessable("user_id wajib diisi", nil)
	}

	_, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperror.NotFound("user tidak ditemukan", err)
	}

	scopeType := valueobject.UserScopeType(strings.TrimSpace(strings.ToLower(req.ScopeType)))
	normalizedValue, err := valueobject.NewUserScopeValue(scopeType, strings.TrimSpace(strings.ToLower(req.ScopeValue)))
	if err != nil {
		return nil, apperror.Unprocessable("scope value tidak valid", nil, err)
	}

	scope, err := entity.NewUserScope(uuid.NewString(), userID, scopeType, normalizedValue)
	if err != nil {
		return nil, apperror.Unprocessable("gagal membuat user scope", nil, err)
	}

	if err := uc.userScopeRepo.Save(ctx, scope); err != nil {
		return nil, apperror.Internal("gagal menyimpan user scope", err)
	}

	return &dto.UserScopeResponse{
		ID:         scope.ID,
		ScopeType:  string(scope.ScopeType),
		ScopeValue: scope.ScopeValue,
	}, nil
}
