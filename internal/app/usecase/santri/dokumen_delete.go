package santriUsecase

import (
	"context"
	"errors"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/port"
	domainerr "sipon-api/internal/domain/errors"
	santriconstant "sipon-api/internal/domain/santri/constant"
	santrirepo "sipon-api/internal/domain/santri/repository"
)

type DokumenDeleteUseCase struct {
	santriRepo        santrirepo.SantriRepository
	santriDokumenRepo santrirepo.SantriDokumenRepository
	fileUploader      port.FileUploader
}

func NewDokumenDeleteUseCase(
	santriRepo santrirepo.SantriRepository,
	santriDokumenRepo santrirepo.SantriDokumenRepository,
	fileUploader port.FileUploader,
) *DokumenDeleteUseCase {
	return &DokumenDeleteUseCase{santriRepo: santriRepo, santriDokumenRepo: santriDokumenRepo, fileUploader: fileUploader}
}

func (uc *DokumenDeleteUseCase) Execute(ctx context.Context, userID, dokumenID string) error {
	santri, err := uc.santriRepo.FindByUserID(ctx, userID)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == santriconstant.CodeSantriNotFound {
			return apperror.NotFound("santri tidak ditemukan", err)
		}
		return apperror.Internal("gagal mengambil data santri", err)
	}

	dokumen, err := uc.santriDokumenRepo.FindByID(ctx, dokumenID)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == santriconstant.CodeDokumenNotFound {
			return apperror.NotFound("dokumen tidak ditemukan", err)
		}
		return apperror.Internal("gagal mengambil dokumen", err)
	}

	if dokumen.SantriID != santri.ID {
		return apperror.Forbidden("akses dokumen ditolak")
	}

	if err := uc.santriDokumenRepo.Delete(ctx, dokumenID); err != nil {
		return apperror.Internal("gagal menghapus dokumen", err)
	}

	_ = uc.fileUploader.DeleteObject(ctx, dokumen.Key, port.PrivacyPrivate)

	return nil
}
