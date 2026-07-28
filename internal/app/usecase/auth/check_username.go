package authUsecase

import (
	"context"
	"errors"
	"strings"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	domainerr "sipon-api/internal/domain/errors"
	userconstant "sipon-api/internal/domain/user/constant"
	userrepo "sipon-api/internal/domain/user/repository"
	uservo "sipon-api/internal/domain/user/valueobject"
)

type CheckUsernameUseCase struct {
	userRepo userrepo.UserRepository
}

func NewCheckUsernameUseCase(userRepo userrepo.UserRepository) *CheckUsernameUseCase {
	return &CheckUsernameUseCase{userRepo: userRepo}
}

// Required — role: any | perm: - | benefit: -
// Execute checks if a username is available (not taken by another user).
func (uc *CheckUsernameUseCase) Execute(ctx context.Context, userID string, username string) (*dto.CheckUsernameResponse, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, apperror.Unprocessable("username tidak boleh kosong", nil)
	}

	_, err := uservo.NewUsername(username)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == userconstant.CodeInvalidUsernameFormat {
			return nil, apperror.Unprocessable(string(de.Code), nil, err)
		}
		return nil, apperror.Unprocessable(string(apperror.CodeUnprocessable), nil)
	}

	existingUser, findErr := uc.userRepo.FindByUsername(ctx, username)
	if findErr != nil {
		var de *domainerr.DomainError
		if errors.As(findErr, &de) {
			if de.Code == userconstant.CodeUserNotFound {
				// Username tidak dipakai siapapun
				return &dto.CheckUsernameResponse{Available: true}, nil
			}
			if de.Code == userconstant.CodeUserQueryFailed {
				return nil, apperror.Internal(string(apperror.CodeInternal), findErr)
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), findErr)
	}

	// Ditemukan user dengan username tersebut
	if existingUser.ID == userID {
		// Username milik user sendiri
		return &dto.CheckUsernameResponse{Available: true}, nil
	}

	return &dto.CheckUsernameResponse{Available: false}, nil
}
