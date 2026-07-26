package authUsecase

import (
	"context"
	"errors"
	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/user/constant"
	usererr "sipon-api/internal/domain/user/constant"
	"sipon-api/internal/domain/user/repository"
	"sipon-api/internal/domain/user/valueobject"
	verificationentity "sipon-api/internal/domain/verification/entity"
	verificationrepo "sipon-api/internal/domain/verification/repository"
	"time"

	"github.com/google/uuid"
)

type RequestIdentityOTPUseCase struct {
	userRepo    repository.UserRepository
	verifRepo   verificationrepo.VerificationRepository
	otpGen      port.OTPGenerator
	emailSender port.EmailSender
	smsSender   port.SMSSender
}

func NewRequestIdentityOTPUseCase(
	userRepo repository.UserRepository,
	verifRepo verificationrepo.VerificationRepository,
	otpGen port.OTPGenerator,
	emailSender port.EmailSender,
	smsSender port.SMSSender,
) *RequestIdentityOTPUseCase {
	return &RequestIdentityOTPUseCase{
		userRepo:    userRepo,
		verifRepo:   verifRepo,
		otpGen:      otpGen,
		emailSender: emailSender,
		smsSender:   smsSender,
	}
}

// Required — role: any | perm: - | benefit: -
func (uc *RequestIdentityOTPUseCase) Execute(ctx context.Context, req dto.RequestIdentityOTPRequest) (*dto.RequestIdentityOTPResponse, error) {
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
		return nil, apperror.Conflict(string(apperror.CodeConflict))
	}

	otpCode, err := uc.otpGen.Generate()
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	verifCode, err := verificationentity.NewVerificationCode(
		uuid.NewString(), identity.UserID, otpCode,
		verificationPurposeFromIdentifier(identifier.Kind()), 5*time.Minute,
	)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	if err := uc.verifRepo.Save(ctx, verifCode); err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	switch identifier.Kind() {
	case constant.LoginIdentifierEmail:
		if err := uc.emailSender.SendOTP(identity.Value, user.Username.Value(), otpCode); err != nil {
			return nil, apperror.Internal(string(apperror.CodeInternal), err)
		}
	case constant.LoginIdentifierPhone:
		if err := uc.smsSender.SendOTP(identity.Value, otpCode); err != nil {
			return nil, apperror.Internal(string(apperror.CodeInternal), err)
		}
	default:
		return nil, apperror.Unprocessable(string(apperror.CodeUnprocessable), nil)
	}

	return &dto.RequestIdentityOTPResponse{Message: "OTP verifikasi berhasil dikirim"}, nil
}
