package santriUsecase

import (
	"context"
	"errors"

	"sipon-api/internal/app/apperror"
	domainerr "sipon-api/internal/domain/errors"
	santriconstant "sipon-api/internal/domain/santri/constant"
	"sipon-api/internal/domain/santri/entity"
	santrirepo "sipon-api/internal/domain/santri/repository"
)

type DokumenListUseCase struct {
	santriRepo        santrirepo.SantriRepository
	santriDokumenRepo santrirepo.SantriDokumenRepository
}

func NewDokumenListUseCase(
	santriRepo santrirepo.SantriRepository,
	santriDokumenRepo santrirepo.SantriDokumenRepository,
) *DokumenListUseCase {
	return &DokumenListUseCase{santriRepo: santriRepo, santriDokumenRepo: santriDokumenRepo}
}

func (uc *DokumenListUseCase) Execute(ctx context.Context, userID, kind string) ([]DokumenItem, error) {
	santri, err := uc.santriRepo.FindByUserID(ctx, userID)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == santriconstant.CodeSantriNotFound {
			return nil, apperror.NotFound("santri tidak ditemukan", err)
		}
		return nil, apperror.Internal("gagal mengambil data santri", err)
	}

	var dokumenList []*entity.SantriDokumen
	if kind != "" {
		dokumenList, err = uc.santriDokumenRepo.FindBySantriIDAndKind(ctx, santri.ID, kind)
	} else {
		dokumenList, err = uc.santriDokumenRepo.FindBySantriID(ctx, santri.ID)
	}
	if err != nil {
		return nil, apperror.Internal("gagal mengambil dokumen", err)
	}

	items := make([]DokumenItem, 0, len(dokumenList))
	for _, d := range dokumenList {
		items = append(items, mapDokumenToItem(d))
	}
	return items, nil
}
