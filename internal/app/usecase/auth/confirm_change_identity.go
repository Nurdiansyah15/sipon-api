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
	userentity "sipon-api/internal/domain/user/entity"
	userrepo "sipon-api/internal/domain/user/repository"
	uservo "sipon-api/internal/domain/user/valueobject"
	verificationconstant "sipon-api/internal/domain/verification/constant"
	verificationrepo "sipon-api/internal/domain/verification/repository"

	"github.com/google/uuid"
)

type ConfirmChangeIdentityUseCase struct {
	userRepo   userrepo.UserRepository
	verifRepo  verificationrepo.VerificationRepository
	transactor port.Transactor
}

func NewConfirmChangeIdentityUseCase(
	userRepo userrepo.UserRepository,
	verifRepo verificationrepo.VerificationRepository,
	transactor port.Transactor,
) *ConfirmChangeIdentityUseCase {
	return &ConfirmChangeIdentityUseCase{userRepo: userRepo, verifRepo: verifRepo, transactor: transactor}
}

type ConfirmChangeIdentityInput struct {
	UserID string
	Kind   userconstant.LoginIdentifierKind
	OTP    string
}

// Required — role: any | perm: - | benefit: -
func (uc *ConfirmChangeIdentityUseCase) Execute(ctx context.Context, in ConfirmChangeIdentityInput) (*dto.ChangeIdentityResponse, error) {
	// 1. Ambil user berdasarkan UserID (dari JWT)
	user, err := uc.userRepo.FindByID(ctx, in.UserID)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == userconstant.CodeUserNotFound {
			return nil, apperror.NotFound(string(apperror.CodeNotFound), err)
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	// 2. Tentukan purpose berdasarkan jenis identitas
	purpose := changeIdentityPurposeFromKind(in.Kind)
	if purpose == "" {
		return nil, apperror.Unprocessable(string(apperror.CodeUnsupportedIdentifierKind), nil)
	}

	// 3. Ambil kode verifikasi terbaru
	code, err := uc.verifRepo.FindLatestByUserAndPurpose(ctx, user.ID, purpose)
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

	// 4. Verifikasi OTP (domain rules: not expired, not used, correct code)
	if err := code.Verify(in.OTP); err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case verificationconstant.CodeOTPExpired, verificationconstant.CodeOTPUsed, verificationconstant.CodeOTPInvalid:
				return nil, apperror.Unprocessable(string(de.Code), nil, err)
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	// 5. Ambil nilai identitas baru dari kode verifikasi
	if code.NewIdentityValue == nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), nil)
	}
	newValue := *code.NewIdentityValue

	// 6. Update atau buat LoginIdentity
	identity := user.FindLoginIdentityByKind(in.Kind)
	if identity != nil {
		// Identity sudah ada (verified maupun unverified) — update nilainya
		identity.Value = newValue
		identity.MarkVerified()
	} else {
		// Pertama kali menambahkan identitas ini (misal: phone baru sama sekali)
		cred := user.FindCredential(userconstant.CredentialTypeLocal)
		if cred == nil {
			return nil, apperror.Internal(string(apperror.CodeInternal), nil)
		}
		now := time.Now()
		newIdentity, err := userentity.NewLoginIdentity(
			uuid.NewString(), user.ID, cred.ID, in.Kind, newValue, true, &now,
		)
		if err != nil {
			return nil, apperror.Internal(string(apperror.CodeInternal), err)
		}
		cred.AddLoginIdentity(newIdentity)
	}

	// 7. Update field identitas pada User aggregate
	switch in.Kind {
	case userconstant.LoginIdentifierEmail:
		email, err := uservo.NewEmail(newValue)
		if err != nil {
			return nil, apperror.Internal(string(apperror.CodeInternal), err)
		}
		user.Email = email
	case userconstant.LoginIdentifierPhone:
		phone, err := uservo.NewPhoneNumber(newValue)
		if err != nil {
			return nil, apperror.Internal(string(apperror.CodeInternal), err)
		}
		user.PhoneNumber = phone
	}
	user.UpdatedAt = time.Now()

	// 8. Persist semua perubahan dalam satu transaksi
	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.userRepo.Update(txCtx, user); err != nil {
			return err
		}
		return uc.verifRepo.Update(txCtx, code)
	}); err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	label := identityKindLabel(in.Kind)
	return &dto.ChangeIdentityResponse{Message: label + " berhasil diperbarui"}, nil
}
