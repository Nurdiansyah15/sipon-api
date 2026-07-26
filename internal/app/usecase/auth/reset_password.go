package authUsecase

import (
	"context"
	"errors"
	"time"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
	domainerr "sipon-api/internal/domain/errors"
	userconstant "sipon-api/internal/domain/user/constant"
	userrepo "sipon-api/internal/domain/user/repository"
	"sipon-api/internal/domain/user/valueobject"
	verificationconstant "sipon-api/internal/domain/verification/constant"
	verificationrepo "sipon-api/internal/domain/verification/repository"
)

type ResetPasswordUseCase struct {
	userRepo  userrepo.UserRepository
	verifRepo verificationrepo.VerificationRepository
	hasher    port.PasswordHasher
}

func NewResetPasswordUseCase(
	userRepo userrepo.UserRepository,
	verifRepo verificationrepo.VerificationRepository,
	hasher port.PasswordHasher,
) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{
		userRepo:  userRepo,
		verifRepo: verifRepo,
		hasher:    hasher,
	}
}

// Required — role: public | perm: - | benefit: -
func (uc *ResetPasswordUseCase) Execute(ctx context.Context, req dto.ResetPasswordRequest) (*dto.ResetPasswordResponse, error) {
	user, err := uc.userRepo.FindByIdentity(ctx, userconstant.LoginIdentifierEmail, req.Email)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case userconstant.CodeLoginIdentityNotFound, userconstant.CodeUserNotFound:
				return nil, apperror.NotFound(string(apperror.CodeNotFound), err)
			case userconstant.CodeUserQueryFailed:
				return nil, apperror.Internal(string(apperror.CodeInternal), err)
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	code, err := uc.verifRepo.FindLatestByUserAndPurpose(ctx, user.ID, verificationconstant.PurposeResetPassword)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case verificationconstant.CodeVerificationNotFound:
				return nil, apperror.NotFound(string(apperror.CodeNotFound), err)
			case verificationconstant.CodeVerificationQueryFailed:
				return nil, apperror.Internal(string(apperror.CodeInternal), err)
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	if err := code.Verify(req.Token); err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case verificationconstant.CodeOTPExpired, verificationconstant.CodeOTPUsed, verificationconstant.CodeOTPInvalid:
				return nil, apperror.Unprocessable(string(de.Code), nil, err)
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	newPlain, err := valueobject.NewPlainPassword(req.Password)
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

	localCred := user.FindCredential(userconstant.CredentialTypeLocal)
	if localCred == nil {
		return nil, apperror.Forbidden(string(userconstant.CodeUserNoLocalCredential))
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
	user.ResetFailedAttempts()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	if err := uc.verifRepo.Update(ctx, code); err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	return &dto.ResetPasswordResponse{Message: "password berhasil direset"}, nil
}
