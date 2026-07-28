package authUsecase

import (
	"context"
	"errors"
	"strings"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
	domainerr "sipon-api/internal/domain/errors"
	userconstant "sipon-api/internal/domain/user/constant"
	userrepo "sipon-api/internal/domain/user/repository"
)

type AvatarConfirmUseCase struct {
	userRepo      userrepo.UserRepository
	transactor    port.Transactor
	fileUploader  port.FileUploader
}

func NewAvatarConfirmUseCase(
	userRepo userrepo.UserRepository,
	transactor port.Transactor,
	fileUploader port.FileUploader,
) *AvatarConfirmUseCase {
	return &AvatarConfirmUseCase{
		userRepo:     userRepo,
		transactor:   transactor,
		fileUploader: fileUploader,
	}
}

// Required — role: any | perm: - | benefit: -
func (uc *AvatarConfirmUseCase) Execute(ctx context.Context, userID string, key string) (*dto.AvatarConfirmResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, apperror.Unauthorized(string(apperror.CodeUnauthorized))
	}

	normalizedKey := uc.fileUploader.KeyFromURL(key)
	if normalizedKey == "" {
		return nil, apperror.Unprocessable(string(apperror.CodeInvalidMediaKey), nil)
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

	oldKey := user.AvatarKey
	user.AvatarKey = &normalizedKey

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return uc.userRepo.Update(txCtx, user)
	}); err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	_ = uc.fileUploader.ConfirmUpload(ctx, normalizedKey)

	if oldKey != nil && *oldKey != normalizedKey {
		markAvatarDeleted(ctx, uc.fileUploader, oldKey)
	}

	return &dto.AvatarConfirmResponse{
		AvatarURL: uc.fileUploader.PublicURL(normalizedKey),
	}, nil
}
