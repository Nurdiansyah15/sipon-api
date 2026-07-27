package usermanagement

import (
	"context"
	"strings"

	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
	userrepo "sipon-api/internal/domain/user/repository"
)

type GetUserUseCase struct {
	userRepo  userrepo.UserRepository
	readModel port.UserQueryReadModel
}

func NewGetUserUseCase(userRepo userrepo.UserRepository, readModel port.UserQueryReadModel) *GetUserUseCase {
	return &GetUserUseCase{userRepo: userRepo, readModel: readModel}
}

// Required — perm: manage_users
func (uc *GetUserUseCase) Execute(ctx context.Context, userID string) (*dto.UserManagementResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		return nil, mapUserDomainError(err)
	}
	return buildUserManagementResponse(ctx, uc.readModel, user, true)
}