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

type AvatarDeleteUseCase struct {
	userRepo     userrepo.UserRepository
	fileUploader port.FileUploader
}

func NewAvatarDeleteUseCase(
	userRepo userrepo.UserRepository,
	fileUploader port.FileUploader,
) *AvatarDeleteUseCase {
	return &AvatarDeleteUseCase{
		userRepo:     userRepo,
		fileUploader: fileUploader,
	}
}

// Required — role: any | perm: - | benefit: -
func (uc *AvatarDeleteUseCase) Execute(ctx context.Context, userID string) (*dto.ChangeIdentityResponse, error) {
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

	oldKey := user.AvatarKey
	user.AvatarKey = nil

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	markAvatarDeleted(ctx, uc.fileUploader, oldKey)

	return &dto.ChangeIdentityResponse{Message: "avatar berhasil dihapus"}, nil
}
