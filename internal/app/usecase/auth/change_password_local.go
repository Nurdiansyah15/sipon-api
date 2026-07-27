package authUsecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
	domainerr "sipon-api/internal/domain/errors"
	userconstant "sipon-api/internal/domain/user/constant"
	userrepo "sipon-api/internal/domain/user/repository"
	"sipon-api/internal/domain/user/valueobject"
)

type ChangePasswordLocalUseCase struct {
	userRepo userrepo.UserRepository
	hasher   port.PasswordHasher
}

func NewChangePasswordLocalUseCase(
	userRepo userrepo.UserRepository,
	hasher port.PasswordHasher,
) *ChangePasswordLocalUseCase {
	return &ChangePasswordLocalUseCase{userRepo: userRepo, hasher: hasher}
}

// Required — role: any | perm: - | benefit: -
func (uc *ChangePasswordLocalUseCase) Execute(ctx context.Context, userID string, req dto.ChangePasswordLocalRequest) (*dto.ChangePasswordLocalResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, apperror.Unauthorized(string(apperror.CodeUnauthorized))
	}

	if req.CurrentPassword == req.NewPassword {
		return nil, apperror.Unprocessable(string(apperror.CodePasswordSameAsCurrent), nil)
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

	localCred := user.FindCredential(userconstant.CredentialTypeLocal)
	if !user.HasLocalPassword() {
		return nil, apperror.Forbidden(string(userconstant.CodeUserNoLocalCredential))
	}

	if err := uc.hasher.Verify(localCred.SecretHash.Value(), req.CurrentPassword); err != nil {
		return nil, apperror.Unprocessable(string(apperror.CodeInvalidCurrentPassword), nil, err)
	}

	hashedStr, err := uc.hasher.Hash(newPlain.Value())
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	newHashed, err := valueobject.NewHashedPassword(hashedStr)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	now := time.Now()
	localCred.SecretHash = &newHashed
	localCred.LastChangedAt = &now
	localCred.UpdatedAt = now
	user.UpdatedAt = now

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	return &dto.ChangePasswordLocalResponse{Message: "password berhasil diubah"}, nil
}
