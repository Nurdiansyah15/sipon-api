package authUsecase

import (
	"context"
	"errors"
	"strings"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
	domainerr "sipon-api/internal/domain/errors"
	userconstant "sipon-api/internal/domain/user/constant"
	userrepo "sipon-api/internal/domain/user/repository"
	"sipon-api/internal/domain/user/valueobject"
)

type SetPasswordLocalUseCase struct {
	userRepo userrepo.UserRepository
	hasher   port.PasswordHasher
}

func NewSetPasswordLocalUseCase(
	userRepo userrepo.UserRepository,
	hasher port.PasswordHasher,
) *SetPasswordLocalUseCase {
	return &SetPasswordLocalUseCase{userRepo: userRepo, hasher: hasher}
}

// Required — role: any | perm: - | benefit: -
// Set password lokal untuk pertama kali (akun social-login-only yang belum
// pernah punya password) — beda dengan ChangePasswordLocalUseCase yang
// mewajibkan verifikasi current_password.
func (uc *SetPasswordLocalUseCase) Execute(ctx context.Context, userID string, req dto.SetPasswordLocalRequest) (*dto.SetPasswordLocalResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, apperror.Unauthorized(string(apperror.CodeUnauthorized))
	}

	newPlain, err := valueobject.NewPlainPassword(req.NewPassword)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case userconstant.CodePasswordTooShort, userconstant.CodePasswordMustHaveUppercase, userconstant.CodePasswordMustHaveDigit:
				return nil, apperror.Unprocessable(string(de.Code), nil, err)
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
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

	hashedStr, err := uc.hasher.Hash(newPlain.Value())
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	newHashed, err := valueobject.NewHashedPassword(hashedStr)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	if err := user.SetLocalPassword(newHashed); err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case userconstant.CodeUserAlreadyHasLocalPassword:
				return nil, apperror.Conflict(string(de.Code), err)
			case userconstant.CodeUserNoLocalCredential:
				return nil, apperror.Internal(string(apperror.CodeInternal), err)
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	return &dto.SetPasswordLocalResponse{Message: "password berhasil ditambahkan"}, nil
}
