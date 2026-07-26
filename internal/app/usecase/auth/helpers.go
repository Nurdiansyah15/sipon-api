package authUsecase

import (
	"errors"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/user/constant"
	"sipon-api/internal/domain/user/entity"

	"github.com/google/uuid"
)

// issueTokenPair menerbitkan access+refresh token untuk user yang sudah lolos
// resolusi identity. deviceID diteruskan apa adanya sampai ke token supaya
// sesi bisa di-revoke per-device lewat SessionRevocationStore.
func issueTokenPair(user *entity.User, deviceID string, tokenGen port.TokenGenerator) (*dto.LoginResponse, error) {
	if err := user.EnsureCanLogin(); err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case constant.CodeUserDeleted:
				return nil, apperror.Forbidden(string(constant.CodeUserDeleted))
			case constant.CodeUserBanned:
				return nil, apperror.Forbidden(string(constant.CodeUserBanned))
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	sessionID := uuid.NewString()
	accessToken, err := tokenGen.GenerateAccessToken(user.ID, sessionID, deviceID)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	refreshToken, err := tokenGen.GenerateRefreshToken(user.ID, deviceID)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	emailIdentity := user.FindLoginIdentity(constant.LoginIdentifierEmail, user.Email.Value())
	isEmailVerified := emailIdentity != nil && emailIdentity.IsVerified()
	var phone *string
	var isPhoneVerified bool
	if user.PhoneNumber != nil {
		pn := user.PhoneNumber.Value()
		phone = &pn
		phoneIdentity := user.FindLoginIdentity(constant.LoginIdentifierPhone, pn)
		isPhoneVerified = phoneIdentity != nil && phoneIdentity.IsVerified()
	}

	return &dto.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: dto.UserMe{
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
		},
	}, nil
}
