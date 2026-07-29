package santriUsecase

import (
	"context"
	"errors"

	"sipon-api/internal/app/apperror"
	domainerr "sipon-api/internal/domain/errors"
	santriconstant "sipon-api/internal/domain/santri/constant"
	santrirepo "sipon-api/internal/domain/santri/repository"
	userrepo "sipon-api/internal/domain/user/repository"
)

type UpdateSantriUseCase struct {
	santriRepo santrirepo.SantriRepository
	userRepo   userrepo.UserRepository
}

func NewUpdateSantriUseCase(
	santriRepo santrirepo.SantriRepository,
	userRepo userrepo.UserRepository,
) *UpdateSantriUseCase {
	return &UpdateSantriUseCase{santriRepo: santriRepo, userRepo: userRepo}
}

func (uc *UpdateSantriUseCase) Execute(ctx context.Context, userID string, req UpdateSantriRequest) (*UpdateSantriResponse, error) {
	santri, err := uc.santriRepo.FindByUserID(ctx, userID)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == santriconstant.CodeSantriNotFound {
			return nil, apperror.NotFound("santri tidak ditemukan", err)
		}
		return nil, apperror.Internal("gagal mengambil data santri", err)
	}

	applySantriUpdate(santri, req)
	santri.Update()

	if err := uc.santriRepo.Update(ctx, santri); err != nil {
		return nil, apperror.Internal("gagal menyimpan data santri", err)
	}

	if req.Fullname != nil {
		user, userErr := uc.userRepo.FindByID(ctx, userID)
		if userErr == nil {
			user.Fullname = req.Fullname
			_ = uc.userRepo.Update(ctx, user)
		}
	}

	return &UpdateSantriResponse{Message: "data santri berhasil diperbarui"}, nil
}
