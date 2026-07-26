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
	verificationconstant "sipon-api/internal/domain/verification/constant"
	verificationentity "sipon-api/internal/domain/verification/entity"
	verificationrepo "sipon-api/internal/domain/verification/repository"

	"github.com/google/uuid"
)

type ForgotPasswordUseCase struct {
	userRepo    userrepo.UserRepository
	verifRepo   verificationrepo.VerificationRepository
	otpGen      port.OTPGenerator
	emailSender port.EmailSender
}

func NewForgotPasswordUseCase(
	userRepo userrepo.UserRepository,
	verifRepo verificationrepo.VerificationRepository,
	otpGen port.OTPGenerator,
	emailSender port.EmailSender,
) *ForgotPasswordUseCase {
	return &ForgotPasswordUseCase{
		userRepo:    userRepo,
		verifRepo:   verifRepo,
		otpGen:      otpGen,
		emailSender: emailSender,
	}
}

// Required — role: public | perm: - | benefit: -
func (uc *ForgotPasswordUseCase) Execute(ctx context.Context, req dto.ForgotPasswordRequest) (*dto.ForgotPasswordResponse, error) {
	user, err := uc.userRepo.FindByIdentity(ctx, userconstant.LoginIdentifierEmail, req.Email)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case userconstant.CodeLoginIdentityNotFound, userconstant.CodeUserNotFound:
				// Return generic message to avoid user enumeration
				return &dto.ForgotPasswordResponse{Message: "jika email terdaftar, OTP reset password telah dikirim"}, nil
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	// Forgot password hanya berlaku untuk akun dengan credential local.
	localCredential := user.FindCredential(userconstant.CredentialTypeLocal)
	if localCredential == nil {
		return &dto.ForgotPasswordResponse{Message: "jika email terdaftar, OTP reset password telah dikirim"}, nil
	}

	// Ambil identity email dari credential local, lalu pastikan sudah verified.
	emailIdentity := localCredential.FindLoginIdentity(userconstant.LoginIdentifierEmail, user.Email.Value())
	if emailIdentity == nil {
		return &dto.ForgotPasswordResponse{Message: "jika email terdaftar, OTP reset password telah dikirim"}, nil
	}
	if err := emailIdentity.EnsureVerified(); err != nil {
		return nil, apperror.Forbidden(string(userconstant.CodeLoginIdentityUnverified))
	}

	otpCode, err := uc.otpGen.Generate()
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	verifCode, err := verificationentity.NewVerificationCode(
		uuid.NewString(), user.ID, otpCode,
		verificationconstant.PurposeResetPassword, 15*time.Minute,
	)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	if err := uc.verifRepo.Save(ctx, verifCode); err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	if err := uc.emailSender.SendPasswordResetOTP(user.Email.Value(), user.Username.Value(), otpCode); err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	return &dto.ForgotPasswordResponse{Message: "jika email terdaftar, OTP reset password telah dikirim"}, nil
}
