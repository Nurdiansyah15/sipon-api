package santriUsecase

import (
	"context"
	"errors"

	"sipon-api/internal/app/apperror"
	domainerr "sipon-api/internal/domain/errors"
	santriconstant "sipon-api/internal/domain/santri/constant"
	santrirepo "sipon-api/internal/domain/santri/repository"
)

type DokumenVerifyUseCase struct {
	santriDokumenRepo santrirepo.SantriDokumenRepository
}

func NewDokumenVerifyUseCase(santriDokumenRepo santrirepo.SantriDokumenRepository) *DokumenVerifyUseCase {
	return &DokumenVerifyUseCase{santriDokumenRepo: santriDokumenRepo}
}

func (uc *DokumenVerifyUseCase) Execute(ctx context.Context, dokumenID, verifierID string) error {
	dokumen, err := uc.santriDokumenRepo.FindByID(ctx, dokumenID)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == santriconstant.CodeDokumenNotFound {
			return apperror.NotFound("dokumen tidak ditemukan", err)
		}
		return apperror.Internal("gagal mengambil dokumen", err)
	}

	if err := dokumen.Verify(verifierID); err != nil {
		return apperror.Conflict("dokumen sudah direject", err)
	}

	if err := uc.santriDokumenRepo.Update(ctx, dokumen); err != nil {
		return apperror.Internal("gagal memverifikasi dokumen", err)
	}

	return nil
}
