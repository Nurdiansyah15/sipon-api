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
)

type MeUseCase struct {
	userRepo userrepo.UserRepository
}

func NewMeUseCase(userRepo userrepo.UserRepository) *MeUseCase {
	return &MeUseCase{userRepo: userRepo}
}

// Required — role: any | perm: - | benefit: -
func (uc *MeUseCase) Execute(ctx context.Context, userID string) (*dto.UserMe, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, apperror.Unauthorized(string(apperror.CodeUnauthorized))
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

	credential := user.FindCredential(userconstant.CredentialTypeLocal)
	if credential == nil {
		return nil, apperror.NotFound(string(apperror.CodeNotFound))
	}

	emailIdentity := credential.FindLoginIdentity(userconstant.LoginIdentifierEmail, user.Email.Value())
	if emailIdentity == nil {
		return nil, apperror.NotFound(string(apperror.CodeNotFound))
	}

	isEmailVerified := emailIdentity.IsVerified()

	var phone *string
	var isPhoneVerified bool
	if user.PhoneNumber != nil {
		pn := user.PhoneNumber.Value()
		phone = &pn
		phoneIdentity := user.FindLoginIdentity(userconstant.LoginIdentifierPhone, pn)
		isPhoneVerified = phoneIdentity != nil && phoneIdentity.IsVerified()
	}

	return &dto.UserMe{
		ID:              user.ID,
		Username:        user.Username.Value(),
		Email:           user.Email.Value(),
		IsEmailVerified: isEmailVerified,
		Fullname:        user.Fullname,
		Phone:           phone,
		IsPhoneVerified: isPhoneVerified,
		Status:          string(user.Status),
		CreatedAt:       user.CreatedAt,
		HasPassword:     user.HasLocalPassword(),
	}, nil
}
