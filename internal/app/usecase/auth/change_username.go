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

type ChangeUsernameUseCase struct {
	userRepo userrepo.UserRepository
}

func NewChangeUsernameUseCase(userRepo userrepo.UserRepository) *ChangeUsernameUseCase {
	return &ChangeUsernameUseCase{userRepo: userRepo}
}

// Required — role: any | perm: - | benefit: -
// Execute changes the user's username directly (no OTP required).
func (uc *ChangeUsernameUseCase) Execute(ctx context.Context, userID string, req dto.ChangeUsernameRequest) (*dto.ChangeUsernameResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, apperror.Unauthorized(string(apperror.CodeUnauthorized))
	}

	newUsernameStr := strings.TrimSpace(req.Username)
	newUsername, err := uservo.NewUsername(newUsernameStr)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == userconstant.CodeInvalidUsernameFormat {
			return nil, apperror.Unprocessable(string(de.Code), nil, err)
		}
		return nil, apperror.Unprocessable(string(apperror.CodeUnprocessable), nil)
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case userconstant.CodeUserNotFound:
				return nil, apperror.NotFound(string(apperror.CodeNotFound), err)
			case userconstant.CodeUserQueryFailed:
				return nil, apperror.Internal(string(apperror.CodeInternal), err)
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	// Cek apakah username sama dengan yang sekarang
	if user.Username.Value() == newUsername.Value() {
		return nil, apperror.Unprocessable("username sama dengan yang sekarang", nil)
	}

	// Cek ketersediaan username
	existingUser, findErr := uc.userRepo.FindByUsername(ctx, newUsername.Value())
	if findErr == nil && existingUser.ID != userID {
		return nil, apperror.Conflict("username sudah digunakan")
	}
	if findErr != nil {
		var de *domainerr.DomainError
		if errors.As(findErr, &de) && de.Code != userconstant.CodeUserNotFound {
			return nil, apperror.Internal(string(apperror.CodeInternal), findErr)
		}
		// CodeUserNotFound → username available, lanjut
	}

	// Update username
	user.ChangeUsername(newUsername)
	if err := uc.userRepo.UpdateUsername(ctx, user.ID, newUsername.Value()); err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	return &dto.ChangeUsernameResponse{
		Message:  "username berhasil diubah",
		Username: newUsername.Value(),
	}, nil
}
