package santriUsecase

import (
	"context"
	"errors"
	"fmt"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/port"
	media "sipon-api/internal/app/service/media"
	domainerr "sipon-api/internal/domain/errors"
	santriconstant "sipon-api/internal/domain/santri/constant"
	santrirepo "sipon-api/internal/domain/santri/repository"
)

type DokumenAccessUseCase struct {
	santriRepo        santrirepo.SantriRepository
	santriDokumenRepo santrirepo.SantriDokumenRepository
	fileUploader      port.FileUploader
}

func NewDokumenAccessUseCase(
	santriRepo santrirepo.SantriRepository,
	santriDokumenRepo santrirepo.SantriDokumenRepository,
	fileUploader port.FileUploader,
) *DokumenAccessUseCase {
	return &DokumenAccessUseCase{santriRepo: santriRepo, santriDokumenRepo: santriDokumenRepo, fileUploader: fileUploader}
}

func (uc *DokumenAccessUseCase) Execute(ctx context.Context, userID, dokumenID string) (*DokumenAccessResponse, error) {
	santri, err := uc.santriRepo.FindByUserID(ctx, userID)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == santriconstant.CodeSantriNotFound {
			return nil, apperror.NotFound("santri tidak ditemukan", err)
		}
		return nil, apperror.Internal("gagal mengambil data santri", err)
	}

	dokumen, err := uc.santriDokumenRepo.FindByID(ctx, dokumenID)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == santriconstant.CodeDokumenNotFound {
			return nil, apperror.NotFound("dokumen tidak ditemukan", err)
		}
		return nil, apperror.Internal("gagal mengambil dokumen", err)
	}

	if dokumen.SantriID != santri.ID {
		return nil, apperror.Forbidden("akses dokumen ditolak")
	}

	accessURL, err := uc.fileUploader.GeneratePresignedDownloadURL(ctx, dokumen.Key, port.PrivacyPrivate, media.SantriDokumenAccessTTL)
	if err != nil {
		return nil, apperror.Internal("gagal membuat URL akses", fmt.Errorf("presign download: %w", err))
	}

	return &DokumenAccessResponse{
		AccessURL: accessURL,
		ExpiresIn: int(media.SantriDokumenAccessTTL.Seconds()),
	}, nil
}
