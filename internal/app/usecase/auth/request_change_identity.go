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
	uservo "sipon-api/internal/domain/user/valueobject"
	verificationentity "sipon-api/internal/domain/verification/entity"
	verificationrepo "sipon-api/internal/domain/verification/repository"

	"github.com/google/uuid"
)

type RequestChangeIdentityUseCase struct {
	userRepo    userrepo.UserRepository
	verifRepo   verificationrepo.VerificationRepository
	otpGen      port.OTPGenerator
	emailSender port.EmailSender
	smsSender   port.SMSSender
}

func NewRequestChangeIdentityUseCase(
	userRepo userrepo.UserRepository,
	verifRepo verificationrepo.VerificationRepository,
	otpGen port.OTPGenerator,
	emailSender port.EmailSender,
	smsSender port.SMSSender,
) *RequestChangeIdentityUseCase {
	return &RequestChangeIdentityUseCase{
		userRepo:    userRepo,
		verifRepo:   verifRepo,
		otpGen:      otpGen,
		emailSender: emailSender,
		smsSender:   smsSender,
	}
}

type RequestChangeIdentityInput struct {
	UserID   string
	Kind     userconstant.LoginIdentifierKind
	NewValue string
}

// Required — role: any | perm: - | benefit: -
func (uc *RequestChangeIdentityUseCase) Execute(ctx context.Context, in RequestChangeIdentityInput) (*dto.ChangeIdentityResponse, error) {
	// 1. Validate format dan normalisasi nilai baru
	var normalizedValue string
	switch in.Kind {
	case userconstant.LoginIdentifierEmail:
		email, err := uservo.NewEmail(in.NewValue)
		if err != nil {
			var de *domainerr.DomainError
			if errors.As(err, &de) && de.Code == userconstant.CodeInvalidEmailFormat {
				return nil, apperror.Unprocessable(string(de.Code), nil, err)
			}
			return nil, apperror.Unprocessable(string(apperror.CodeUnprocessable), nil)
		}
		normalizedValue = email.Value()
	case userconstant.LoginIdentifierPhone:
		phone, err := uservo.NewPhoneNumber(in.NewValue)
		if err != nil {
			var de *domainerr.DomainError
			if errors.As(err, &de) && de.Code == userconstant.CodeInvalidPhoneNumberFormat {
				return nil, apperror.Unprocessable(string(de.Code), nil, err)
			}
			return nil, apperror.Unprocessable(string(apperror.CodeUnprocessable), nil)
		}
		normalizedValue = phone.Value()
	default:
		return nil, apperror.Unprocessable(string(apperror.CodeUnprocessable), nil)
	}

	// 2. Uniqueness check dengan ownership awareness
	existingUser, findErr := uc.userRepo.FindByIdentity(ctx, in.Kind, normalizedValue)
	if findErr == nil {
		// Ada user yang memiliki identitas ini
		if existingUser.ID != in.UserID {
			return nil, apperror.Conflict(string(apperror.CodeConflict))
		}
		// User sendiri yang punya — cek apakah sudah verified
		identity := existingUser.FindLoginIdentityByKind(in.Kind)
		if identity != nil && identity.IsVerified() {
			return nil, apperror.Conflict(string(apperror.CodeConflict))
		}
		// Unverified + nilai sama → izinkan re-send OTP
	} else {
		var de *domainerr.DomainError
		if errors.As(findErr, &de) {
			if de.Code != userconstant.CodeUserNotFound && de.Code != userconstant.CodeLoginIdentityNotFound {
				return nil, apperror.Internal(string(apperror.CodeInternal), findErr)
			}
		}
		// Not found → nilai baru bebas digunakan
	}

	// 3. Ambil user yang melakukan request
	user, err := uc.userRepo.FindByID(ctx, in.UserID)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == userconstant.CodeUserNotFound {
			return nil, apperror.NotFound(string(apperror.CodeNotFound), err)
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	// 4. Generate OTP
	otpCode, err := uc.otpGen.Generate()
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	// 5. Buat VerificationCode dan simpan nilai identitas baru di dalamnya
	purpose := changeIdentityPurposeFromKind(in.Kind)
	verifCode, err := verificationentity.NewVerificationCode(
		uuid.NewString(), user.ID, otpCode, purpose, 5*time.Minute,
	)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	verifCode.SetNewIdentityValue(normalizedValue)

	if err := uc.verifRepo.Save(ctx, verifCode); err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	// 6. Kirim OTP ke alamat BARU
	switch in.Kind {
	case userconstant.LoginIdentifierEmail:
		if err := uc.emailSender.SendOTP(normalizedValue, user.Username.Value(), otpCode); err != nil {
			return nil, apperror.Internal(string(apperror.CodeInternal), err)
		}
	case userconstant.LoginIdentifierPhone:
		if err := uc.smsSender.SendOTP(normalizedValue, otpCode); err != nil {
			return nil, apperror.Internal(string(apperror.CodeInternal), err)
		}
	}

	label := identityKindLabel(in.Kind)
	return &dto.ChangeIdentityResponse{Message: "OTP berhasil dikirim ke " + label + " baru"}, nil
}
