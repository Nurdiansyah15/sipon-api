package santriUsecase

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/google/uuid"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/port"
	media "sipon-api/internal/app/service/media"
	santriconstant "sipon-api/internal/domain/santri/constant"
)

type DokumenPresignUseCase struct {
	fileUploader port.FileUploader
}

func NewDokumenPresignUseCase(fileUploader port.FileUploader) *DokumenPresignUseCase {
	return &DokumenPresignUseCase{fileUploader: fileUploader}
}

func (uc *DokumenPresignUseCase) Execute(ctx context.Context, req DokumenPresignRequest) (*DokumenPresignResponse, error) {
	ct := strings.TrimSpace(strings.ToLower(req.ContentType))
	if !santriconstant.AllowedContentTypes[ct] {
		return nil, apperror.Unprocessable("tipe konten tidak diizinkan", nil)
	}

	kind, err := parseDokumenKind(req.Kind)
	if err != nil {
		return nil, apperror.Unprocessable(err.Error(), nil)
	}

	ext := ".pdf"
	if strings.HasPrefix(ct, "image/jpeg") || strings.HasPrefix(ct, "image/jpg") {
		ext = ".jpg"
	} else if strings.HasPrefix(ct, "image/png") {
		ext = ".png"
	}

	objectName := path.Join(string(media.ObjectPathSantriDokumen), string(kind), uuid.NewString()+ext)

	presignURL, key, _, err := uc.fileUploader.RequestUpload(ctx, objectName, ct, media.SantriDokumenPresignUploadExpiry, port.PrivacyPrivate)
	if err != nil {
		return nil, apperror.Internal("gagal membuat presigned URL", fmt.Errorf("presign failed: %w", err))
	}

	return &DokumenPresignResponse{
		PresignURL: presignURL,
		Key:        key,
		ExpiresIn:  int(media.SantriDokumenPresignUploadExpiry.Seconds()),
	}, nil
}
