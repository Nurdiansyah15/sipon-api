package authUsecase

import (
	"context"
	"errors"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	domainerr "sipon-api/internal/domain/errors"
	usererr "sipon-api/internal/domain/user/constant"
	"sipon-api/internal/domain/user/repository"
	"sipon-api/internal/domain/user/valueobject"
	verificationConstant "sipon-api/internal/domain/verification/constant"
	verificationrepo "sipon-api/internal/domain/verification/repository"
)

type VerifyIdentityOTPUseCase struct {
	userRepo  repository.UserRepository
	verifRepo verificationrepo.VerificationRepository
}

func NewVerifyIdentityOTPUseCase(
	userRepo repository.UserRepository,
	verifRepo verificationrepo.VerificationRepository,
) *VerifyIdentityOTPUseCase {
	return &VerifyIdentityOTPUseCase{
		userRepo:  userRepo,
		verifRepo: verifRepo,
	}
}

// Required — role: any | perm: - | benefit: -
func (uc *VerifyIdentityOTPUseCase) Execute(ctx context.Context, req dto.VerifyIdentityOTPRequest) (*dto.VerifyIdentityOTPResponse, error) {
	identifier, err := valueobject.NewLoginIdentifier(req.Identifier)
	if err != nil {
		return nil, apperror.Unprocessable(string(apperror.CodeUnprocessable), nil)
	}
	user, err := uc.userRepo.FindByLoginIdentifier(ctx, identifier)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case usererr.CodeLoginIdentityNotFound, usererr.CodeUserNotFound:
				return nil, apperror.NotFound(string(apperror.CodeNotFound), err)
			case usererr.CodeUserQueryFailed:
				return nil, apperror.Internal(string(apperror.CodeInternal), err)
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	identity := user.FindLoginIdentity(identifier.Kind(), identifier.Value())
	if identity == nil {
		return nil, apperror.NotFound(string(apperror.CodeNotFound))
	}
	if identity.IsVerified() {
		return &dto.VerifyIdentityOTPResponse{Message: "identity sudah terverifikasi"}, nil
	}

	purpose := verificationPurposeFromIdentifier(identifier.Kind())
	if purpose == "" {
		return nil, apperror.Unprocessable(string(apperror.CodeUnsupportedIdentifierKind), nil)
	}

	code, err := uc.verifRepo.FindLatestByUserAndPurpose(ctx, identity.UserID, purpose)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case verificationConstant.CodeVerificationNotFound:
				return nil, apperror.NotFound(string(apperror.CodeNotFound), err)
			case verificationConstant.CodeVerificationQueryFailed:
				return nil, apperror.Internal(string(apperror.CodeInternal), err)
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	if err := code.Verify(req.OTP); err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case verificationConstant.CodeOTPExpired, verificationConstant.CodeOTPUsed, verificationConstant.CodeOTPInvalid:
				return nil, apperror.Unprocessable(string(de.Code), nil, err)
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	identity.MarkVerified()
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	if err := uc.verifRepo.Update(ctx, code); err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	return &dto.VerifyIdentityOTPResponse{Message: "identity berhasil diverifikasi"}, nil
}
