package santriUsecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/port"
	domainerr "sipon-api/internal/domain/errors"
	santriconstant "sipon-api/internal/domain/santri/constant"
	"sipon-api/internal/domain/santri/entity"
	santrirepo "sipon-api/internal/domain/santri/repository"
)

type DokumenConfirmUseCase struct {
	santriRepo        santrirepo.SantriRepository
	santriDokumenRepo santrirepo.SantriDokumenRepository
	fileUploader      port.FileUploader
}

func NewDokumenConfirmUseCase(
	santriRepo santrirepo.SantriRepository,
	santriDokumenRepo santrirepo.SantriDokumenRepository,
	fileUploader port.FileUploader,
) *DokumenConfirmUseCase {
	return &DokumenConfirmUseCase{
		santriRepo:        santriRepo,
		santriDokumenRepo: santriDokumenRepo,
		fileUploader:      fileUploader,
	}
}

func (uc *DokumenConfirmUseCase) Execute(ctx context.Context, userID string, req DokumenConfirmRequest) (*DokumenConfirmResponse, error) {
	santri, err := uc.santriRepo.FindByUserID(ctx, userID)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == santriconstant.CodeSantriNotFound {
			return nil, apperror.NotFound("santri tidak ditemukan", err)
		}
		return nil, apperror.Internal("gagal mengambil data santri", err)
	}

	kind, err := parseDokumenKind(req.Kind)
	if err != nil {
		return nil, apperror.Unprocessable(err.Error(), nil)
	}

	normalizedKey := uc.fileUploader.KeyFromURL(req.Key)
	if normalizedKey == "" {
		return nil, apperror.Unprocessable("key dokumen tidak valid", nil)
	}

	dokumen, err := entity.NewSantriDokumen(uuid.NewString(), santri.ID, kind, normalizedKey)
	if err != nil {
		return nil, apperror.Unprocessable("gagal membuat dokumen", fmt.Errorf("new dokumen: %w", err))
	}

	dokumen.OriginalFilename = req.OriginalFilename
	dokumen.MimeType = req.MimeType
	dokumen.Size = req.Size

	if err := uc.santriDokumenRepo.Save(ctx, dokumen); err != nil {
		return nil, apperror.Internal("gagal menyimpan dokumen", err)
	}

	confirmDokumenMediaKey(ctx, uc.fileUploader, normalizedKey)

	return &DokumenConfirmResponse{
		ID:        dokumen.ID,
		Kind:      string(dokumen.Kind),
		Key:       dokumen.Key,
		Status:    string(dokumen.Status),
		CreatedAt: dokumen.CreatedAt,
	}, nil
}
