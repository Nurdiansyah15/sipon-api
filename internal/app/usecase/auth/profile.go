package authUsecase

import (
	"context"
	"errors"
	"strings"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
	"sipon-api/internal/app/service/principal"
	domainerr "sipon-api/internal/domain/errors"
	userconstant "sipon-api/internal/domain/user/constant"
	userrepo "sipon-api/internal/domain/user/repository"
)

type GetProfileUseCase struct {
	userRepo     userrepo.UserRepository
	fileUploader port.FileUploader
}

func NewGetProfileUseCase(userRepo userrepo.UserRepository, fileUploader port.FileUploader) *GetProfileUseCase {
	return &GetProfileUseCase{userRepo: userRepo, fileUploader: fileUploader}
}

// Required — role: any | perm: - | benefit: -
// Execute merges profile (UserMe) + session (roles+permissions) in one call.
func (uc *GetProfileUseCase) Execute(ctx context.Context, userID string, p *principal.Principal) (*dto.ProfileResponse, error) {
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

	roles := make([]dto.SessionRole, 0, len(p.Roles))
	for _, r := range p.Roles {
		roles = append(roles, dto.SessionRole{
			Name:      r.Name,
			RoleType:  r.RoleType,
			ScopeType: r.ScopeType,
			ScopeID:   r.ScopeID,
		})
	}

	permissions := make([]dto.SessionPermission, 0, len(p.Permissions))
	for _, perm := range p.Permissions {
		permissions = append(permissions, dto.SessionPermission{
			Key:   perm.Key,
			Scope: perm.Scope,
		})
	}

	scopes := make([]dto.SessionUserScope, 0, len(p.Scopes))
	for _, s := range p.Scopes {
		scopes = append(scopes, dto.SessionUserScope{
			ScopeType:  s.ScopeType,
			ScopeValue: s.ScopeValue,
		})
	}

	return &dto.ProfileResponse{
		ID:              user.ID,
		Username:        user.Username.Value(),
		Fullname:        user.Fullname,
		Email:           user.Email.Value(),
		IsEmailVerified: isEmailVerified,
		Phone:           phone,
		IsPhoneVerified: isPhoneVerified,
		Status:          string(user.Status),
		HasPassword:     user.HasLocalPassword(),
		CreatedAt:       user.CreatedAt,
		AvatarURL:       resolveAvatarURL(uc.fileUploader, user.AvatarKey),
		Roles:           roles,
		Permissions:     permissions,
		Scopes:          scopes,
	}, nil
}
