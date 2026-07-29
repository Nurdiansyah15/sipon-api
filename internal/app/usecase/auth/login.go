package authUsecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/user/constant"
	usererr "sipon-api/internal/domain/user/constant"
	userrepo "sipon-api/internal/domain/user/repository"
	"sipon-api/internal/domain/user/valueobject"
)

type LoginUseCase struct {
	userRepo userrepo.UserRepository
	hasher   port.PasswordHasher
	tokenGen port.TokenGenerator
}

func NewLoginUseCase(
	userRepo userrepo.UserRepository,
	hasher port.PasswordHasher,
	tokenGen port.TokenGenerator,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo: userRepo,
		hasher:   hasher,
		tokenGen: tokenGen,
	}
}

// Required — role: public | perm: - | benefit: -
func (uc *LoginUseCase) Execute(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	identifier, err := valueobject.NewLoginIdentifier(req.Identifier)
	if err != nil {
		return nil, apperror.Unauthorized(string(apperror.CodeUnauthorized))
	}

	user, err := uc.userRepo.FindByLoginIdentifier(ctx, identifier)
	if err != nil {
		if identifier.Kind() == usererr.LoginIdentifierNIS {
			user, err = uc.userRepo.FindByIdentity(ctx, usererr.LoginIdentifierUsername, identifier.Value())
		}
	}
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case usererr.CodeLoginIdentityNotFound, usererr.CodeUserNotFound:
				return nil, apperror.Unauthorized(string(apperror.CodeUnauthorized), err)
			case usererr.CodeUserQueryFailed:
				return nil, apperror.Internal(string(apperror.CodeInternal), err)
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	if err := user.EnsureNotLockedOut(); err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == usererr.CodeUserLockedOut {
			return nil, apperror.TooManyRequests(string(usererr.CodeUserLockedOut), err)
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	identity := user.FindLoginIdentity(identifier.Kind(), identifier.Value())
	if identity == nil {
		return nil, apperror.Unauthorized(string(apperror.CodeUnauthorized))
	}

	cred := user.FindCredential(constant.CredentialTypeLocal)
	if cred == nil {
		return nil, apperror.Unauthorized(string(apperror.CodeUnauthorized))
	}

	if cred.SecretHash == nil {
		return nil, apperror.Unauthorized(string(apperror.CodeUnauthorized))
	}
	if err := uc.hasher.Verify(cred.SecretHash.Value(), req.Password); err != nil {
		user.IncrementFailedAttempts()
		_ = uc.userRepo.Update(ctx, user)
		return nil, apperror.Unauthorized(string(apperror.CodeUnauthorized))
	}

	if err := user.EnsureCanLogin(); err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case usererr.CodeUserDeleted:
				return nil, apperror.Forbidden(string(usererr.CodeUserDeleted))
			case usererr.CodeUserBanned:
				return nil, apperror.Forbidden(string(usererr.CodeUserBanned))
			case usererr.CodeUserNoCredential:
				return nil, apperror.Unauthorized(string(apperror.CodeUnauthorized))
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	emailIdentityCheck := user.FindLoginIdentity(constant.LoginIdentifierEmail, user.Email.Value())

	user.ResetFailedAttempts()
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	sessionID := uuid.NewString()
	accessToken, err := uc.tokenGen.GenerateAccessToken(user.ID, sessionID, req.DeviceID)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	refreshToken, err := uc.tokenGen.GenerateRefreshToken(user.ID, req.DeviceID)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	isEmailVerified := emailIdentityCheck.IsVerified()
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
