package santriUsecase

import (
	"context"
	"errors"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/port"
	domainerr "sipon-api/internal/domain/errors"
	santriconstant "sipon-api/internal/domain/santri/constant"
	santrirepo "sipon-api/internal/domain/santri/repository"
	userconstant "sipon-api/internal/domain/user/constant"
	userrepo "sipon-api/internal/domain/user/repository"
)

type GetSantriUseCase struct {
	santriRepo   santrirepo.SantriRepository
	userRepo     userrepo.UserRepository
	fileUploader port.FileUploader
}

func NewGetSantriUseCase(
	santriRepo santrirepo.SantriRepository,
	userRepo userrepo.UserRepository,
	fileUploader port.FileUploader,
) *GetSantriUseCase {
	return &GetSantriUseCase{santriRepo: santriRepo, userRepo: userRepo, fileUploader: fileUploader}
}

func (uc *GetSantriUseCase) Execute(ctx context.Context, userID string) (*GetSantriResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == userconstant.CodeUserNotFound {
			return nil, apperror.NotFound("user not found", err)
		}
		return nil, apperror.Internal("gagal mengambil user", err)
	}

	santri, err := uc.santriRepo.FindByUserID(ctx, userID)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == santriconstant.CodeSantriNotFound {
			return nil, apperror.NotFound("santri not found", err)
		}
		return nil, apperror.Internal("gagal mengambil data santri", err)
	}

	var avatarURL *string
	if user.AvatarKey != nil && *user.AvatarKey != "" {
		url := uc.fileUploader.PublicURL(*user.AvatarKey)
		avatarURL = &url
	}

	var phone *string
	if user.PhoneNumber != nil {
		p := user.PhoneNumber.Value()
		phone = &p
	}

	email := user.Email.Value()
	fullname := user.Fullname

	return mapSantriToResponse(santri, fullname, &email, phone, avatarURL), nil
}
