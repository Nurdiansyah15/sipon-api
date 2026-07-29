package santriUsecase

import (
	"context"
	"errors"

	"sipon-api/internal/app/apperror"
	domainerr "sipon-api/internal/domain/errors"
	santriconstant "sipon-api/internal/domain/santri/constant"
	santrirepo "sipon-api/internal/domain/santri/repository"
)

type DokumenRejectUseCase struct {
	santriDokumenRepo santrirepo.SantriDokumenRepository
}

func NewDokumenRejectUseCase(santriDokumenRepo santrirepo.SantriDokumenRepository) *DokumenRejectUseCase {
	return &DokumenRejectUseCase{santriDokumenRepo: santriDokumenRepo}
}

func (uc *DokumenRejectUseCase) Execute(ctx context.Context, dokumenID, verifierID string, notes *string) error {
	dokumen, err := uc.santriDokumenRepo.FindByID(ctx, dokumenID)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == santriconstant.CodeDokumenNotFound {
			return apperror.NotFound("dokumen tidak ditemukan", err)
		}
		return apperror.Internal("gagal mengambil dokumen", err)
	}

	if err := dokumen.Reject(verifierID, notes); err != nil {
		return apperror.Internal("gagal mereject dokumen", err)
	}

	if err := uc.santriDokumenRepo.Update(ctx, dokumen); err != nil {
		return apperror.Internal("gagal menyimpan status reject", err)
	}

	return nil
}
