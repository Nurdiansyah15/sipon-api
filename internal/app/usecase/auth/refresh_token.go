package authUsecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
	domainerr "sipon-api/internal/domain/errors"
	userconstant "sipon-api/internal/domain/user/constant"
	userrepo "sipon-api/internal/domain/user/repository"
)

type RefreshTokenUseCase struct {
	userRepo        userrepo.UserRepository
	tokenGen        port.TokenGenerator
	revocationStore port.SessionRevocationStore
}

func NewRefreshTokenUseCase(
	userRepo userrepo.UserRepository,
	tokenGen port.TokenGenerator,
	revocationStore port.SessionRevocationStore,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		userRepo:        userRepo,
		tokenGen:        tokenGen,
		revocationStore: revocationStore,
	}
}

// Required — role: public | perm: - | benefit: -
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, req dto.RefreshTokenRequest) (*dto.LoginResponse, error) {
	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		return nil, apperror.Unprocessable(string(apperror.CodeTokenRequired), nil)
	}

	claims, err := uc.tokenGen.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, apperror.Unauthorized(string(apperror.CodeUnauthorized), err)
	}
	userID := strings.TrimSpace(claims.UserID)
	if userID == "" {
		return nil, apperror.Unauthorized(string(apperror.CodeUnauthorized))
	}

	// Enforce logout-all: refresh token diterbitkan sebelum revoke-all terakhir
	// dianggap tidak valid lagi, meski belum expired. Best-effort — kalau store
	// gagal dibaca (mis. Redis sempat down), jangan blokir refresh yang legit.
	if uc.revocationStore != nil {
		if revokedBefore, revErr := uc.revocationStore.RevokedBefore(ctx, userID); revErr == nil && revokedBefore != nil {
			if claims.IssuedAt.Before(*revokedBefore) {
				return nil, apperror.Unauthorized(string(apperror.CodeUnauthorized))
			}
		}
		// Enforce "logout device lain": refresh token ini terikat ke deviceID
		// (dikirim client saat login) — kalau device itu di-logout, refresh-nya
		// ditolak di sini juga, bukan cuma access token-nya yang expired sendiri.
		if claims.DeviceID != "" {
			if revokedBefore, revErr := uc.revocationStore.DeviceRevokedBefore(ctx, userID, claims.DeviceID); revErr == nil && revokedBefore != nil {
				if claims.IssuedAt.Before(*revokedBefore) {
					return nil, apperror.Unauthorized(string(apperror.CodeUnauthorized))
				}
			}
		}
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case userconstant.CodeUserNotFound:
				return nil, apperror.Unauthorized(string(apperror.CodeUnauthorized), err)
			case userconstant.CodeUserQueryFailed:
				return nil, apperror.Internal(string(apperror.CodeInternal), err)
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	if err := user.EnsureCanLogin(); err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case userconstant.CodeUserDeleted:
				return nil, apperror.Forbidden(string(userconstant.CodeUserDeleted))
			case userconstant.CodeUserBanned:
				return nil, apperror.Forbidden(string(userconstant.CodeUserBanned))
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	sessionID := uuid.NewString()
	// deviceID dibawa terus (carry-forward) dari refresh token lama — client
	// tidak perlu resend device_id di setiap refresh, cukup sekali saat login.
	accessToken, err := uc.tokenGen.GenerateAccessToken(user.ID, sessionID, claims.DeviceID)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	newRefreshToken, err := uc.tokenGen.GenerateRefreshToken(user.ID, claims.DeviceID)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	emailIdentity := user.FindLoginIdentity(userconstant.LoginIdentifierEmail, user.Email.Value())
	isEmailVerified := emailIdentity != nil && emailIdentity.IsVerified()
	var phone *string
	var isPhoneVerified bool
	if user.PhoneNumber != nil {
		pn := user.PhoneNumber.Value()
		phone = &pn
		phoneIdentity := user.FindLoginIdentity(userconstant.LoginIdentifierPhone, pn)
		isPhoneVerified = phoneIdentity != nil && phoneIdentity.IsVerified()
	}

	return &dto.LoginResponse{
		Token:        accessToken,
		RefreshToken: newRefreshToken,
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
