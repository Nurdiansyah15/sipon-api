package authUsecase

import (
	"context"
	"errors"
	"time"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	domainerr "sipon-api/internal/domain/errors"
	userconstant "sipon-api/internal/domain/user/constant"
	userrepo "sipon-api/internal/domain/user/repository"
	uservo "sipon-api/internal/domain/user/valueobject"
)

type UpdateProfileUseCase struct {
	userRepo userrepo.UserRepository
}

func NewUpdateProfileUseCase(userRepo userrepo.UserRepository) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{userRepo: userRepo}
}

// Required — role: any | perm: - | benefit: -
// Execute updates profile fields: fullname, email (unverified only), phone (unverified only).
func (uc *UpdateProfileUseCase) Execute(ctx context.Context, userID string, req dto.UpdateProfileRequest) (*dto.UpdateProfileResponse, error) {
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

	cred := user.FindCredential(userconstant.CredentialTypeLocal)
	if cred == nil {
		return nil, apperror.NotFound(string(apperror.CodeNotFound))
	}

	// ── Update Fullname ─────────────────────────────────────────────────────
	if req.Fullname != nil {
		user.Fullname = req.Fullname
	}

	// ── Update Email (only if unverified) ───────────────────────────────────
	if req.Email != nil {
		currentEmailIdentity := cred.FindLoginIdentity(userconstant.LoginIdentifierEmail, user.Email.Value())
		if currentEmailIdentity != nil && currentEmailIdentity.IsVerified() {
			return nil, apperror.Conflict("email sudah diverifikasi. Gunakan endpoint change-email untuk mengganti email")
		}

		newEmail, err := uservo.NewEmail(*req.Email)
		if err != nil {
			var de *domainerr.DomainError
			if errors.As(err, &de) && de.Code == userconstant.CodeInvalidEmailFormat {
				return nil, apperror.Unprocessable(string(de.Code), nil, err)
			}
			return nil, apperror.Unprocessable(string(apperror.CodeUnprocessable), nil)
		}

		// Check uniqueness (ownership-aware)
		existingUser, findErr := uc.userRepo.FindByIdentity(ctx, userconstant.LoginIdentifierEmail, newEmail.Value())
		if findErr == nil && existingUser.ID != userID {
			return nil, apperror.Conflict("email sudah digunakan")
		}

		user.Email = newEmail
		if currentEmailIdentity != nil {
			currentEmailIdentity.Value = newEmail.Value()
		}
	}

	// ── Update Phone (only if unverified) ───────────────────────────────────
	if req.Phone != nil {
		if user.PhoneNumber != nil {
			pn := user.PhoneNumber.Value()
			currentPhoneIdentity := cred.FindLoginIdentity(userconstant.LoginIdentifierPhone, pn)
			if currentPhoneIdentity != nil && currentPhoneIdentity.IsVerified() {
				return nil, apperror.Conflict("nomor telepon sudah diverifikasi. Gunakan endpoint change-phone untuk mengganti nomor telepon")
			}
		}

		newPhone, err := uservo.NewPhoneNumber(*req.Phone)
		if err != nil {
			var de *domainerr.DomainError
			if errors.As(err, &de) && de.Code == userconstant.CodeInvalidPhoneNumberFormat {
				return nil, apperror.Unprocessable(string(de.Code), nil, err)
			}
			return nil, apperror.Unprocessable(string(apperror.CodeUnprocessable), nil)
		}

		// Check uniqueness (ownership-aware)
		existingUser, findErr := uc.userRepo.FindByIdentity(ctx, userconstant.LoginIdentifierPhone, newPhone.Value())
		if findErr == nil && existingUser.ID != userID {
			return nil, apperror.Conflict("nomor telepon sudah digunakan")
		}

		user.PhoneNumber = newPhone
		if user.PhoneNumber != nil {
			existingPhoneIdentity := cred.FindLoginIdentityByKind(userconstant.LoginIdentifierPhone)
			if existingPhoneIdentity != nil {
				existingPhoneIdentity.Value = newPhone.Value()
			}
		}
	}

	user.UpdatedAt = time.Now()
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	return &dto.UpdateProfileResponse{Message: "profil berhasil diperbarui"}, nil
}
